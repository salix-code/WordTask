# -*- coding: utf-8 -*-
"""
用 ECDICT 的中文释义 / 音标 / 词频标签，给 KET A2 Key 词库做 enrichment。

输入：
  data/wordbooks/ket-a2-key.json      —— KET 原始词表（honorwu/ketwords 抽取）
  data/ecdict/ecdict.csv              —— ECDICT 词典（skywind3000/ECDICT）

输出：
  data/wordbooks/ket-a2-key.enriched.json    —— 每条新增 translation/definition/phonetic/tag/ecdictMatch
  data/wordbooks/ket-a2-key.unmatched.json   —— 未匹配到的词，供人工补齐
  控制台：匹配率统计

匹配策略（按优先级，命中即停）：
  1. term 全小写精确匹配
  2. baseTerm 全小写精确匹配
  3. acceptedSpellings 里任一别名精确匹配（英/美变体，如 colour/color）
  4. 去掉空格、连字符后精确匹配
  5. 去掉末尾感叹号/问号（Yeah! / congratulations!）
  6. 去掉带词性/方言标注括号尾缀，如 "movie (n) (Am Eng)" -> "movie"
  7. 处理可选字母括号：
       - "blond(e)"   -> 先试 "blonde"，再试 "blond"
       - "gram(me)"   -> 先试 "gramme"，再试 "gram"
       - "yog(h)urt"  -> 先试 "yoghurt"，再试 "yogurt"
       - "photo(graph)" -> 先试 "photograph"，再试 "photo"
  8. 处理可选后缀括号："mobile (phone)" -> 先试 "mobile phone"，再试 "mobile"
  9. 短语取第一个实词（仅 learningTarget=recognize），例如 "get on" -> "get"

KET 词表中有短语（a few / get on / take off）和合写词（a/an），合理地未命中属正常，记入 unmatched 供人工处理。
"""

from __future__ import annotations

import csv
import json
import re
import sys
from pathlib import Path
from typing import Dict, List, Optional, Tuple

# Windows 控制台输出 UTF-8
sys.stdout.reconfigure(encoding="utf-8")

ROOT = Path(__file__).resolve().parent.parent
KET_INPUT = ROOT / "data" / "wordbooks" / "ket-a2-key.json"
ECDICT_CSV = ROOT / "data" / "ecdict" / "ecdict.csv"
OUT_ENRICHED = ROOT / "data" / "wordbooks" / "ket-a2-key.enriched.json"
OUT_UNMATCHED = ROOT / "data" / "wordbooks" / "ket-a2-key.unmatched.json"


# -------------------- 1. 加载 ECDICT --------------------

def _clean_text(s: str) -> str:
    """ECDICT 里用字面 \\n / \\r\\n 做换行分隔，这里还原成真正的换行。"""
    if not s:
        return ""
    return (
        s.replace("\\r\\n", "\n")
         .replace("\\n", "\n")
         .replace("\r\n", "\n")
         .strip()
    )


def load_ecdict(path: Path) -> Dict[str, dict]:
    """
    ECDICT CSV 列: word, phonetic, definition, translation, pos, collins, oxford, tag, bnc, frq, exchange, detail, audio
    返回 { lowercase_word: {phonetic, definition, translation, pos, tag, bnc, frq, exchange} }
    """
    # 用较大缓冲区，加快读取
    csv.field_size_limit(10 * 1024 * 1024)
    d: Dict[str, dict] = {}
    with path.open("r", encoding="utf-8", newline="") as f:
        reader = csv.DictReader(f)
        for row in reader:
            w = (row.get("word") or "").strip().lower()
            if not w:
                continue
            d[w] = {
                "phonetic": (row.get("phonetic") or "").strip(),
                "definition": _clean_text(row.get("definition") or ""),
                "translation": _clean_text(row.get("translation") or ""),
                "pos": (row.get("pos") or "").strip(),
                "tag": (row.get("tag") or "").strip(),
                "bnc": (row.get("bnc") or "").strip(),
                "frq": (row.get("frq") or "").strip(),
                "exchange": (row.get("exchange") or "").strip(),
            }
    return d


# -------------------- 2. 匹配策略 --------------------

_SPACE_RE = re.compile(r"[\s\-]+")
# 尾缀标注，例如 "movie (n) (Am Eng)" / "train (transitive and intransitive)"
_TRAILING_PAREN_RE = re.compile(r"\s*\([^)]*\)\s*$")
# 行内可选字母括号，例如 "blond(e)" / "gram(me)" / "yog(h)urt" / "photo(graph)"
_INLINE_PAREN_RE = re.compile(r"\(([^)]*)\)")
# 行末感叹号/问号
_TRAILING_PUNCT_RE = re.compile(r"[!?.]+$")


def _norm(s: str) -> str:
    return s.strip().lower()


def _gen_candidates(term: str) -> List[str]:
    """
    根据原始 term 生成一组候选查询 key（已去重、已小写），按优先级排序。
    处理：末尾标点、尾缀标注括号、行内可选字母括号、空格/连字符压缩。
    """
    out: List[str] = []
    seen = set()

    def push(x: str) -> None:
        x = x.strip().lower()
        if x and x not in seen:
            seen.add(x)
            out.append(x)

    base = term.strip()

    # 1) 去末尾标点
    no_punct = _TRAILING_PUNCT_RE.sub("", base)
    push(no_punct)

    # 2) 去末尾括号标注（可能有多个：`movie (n) (Am Eng)`）
    stripped = no_punct
    while True:
        new_s = _TRAILING_PAREN_RE.sub("", stripped).strip()
        if new_s == stripped:
            break
        stripped = new_s
    push(stripped)

    # 3) 处理行内可选字母 / 可选后缀括号
    #    "blond(e)"    -> ["blonde", "blond"]
    #    "mobile (phone)" -> ["mobile phone", "mobile"]
    if "(" in stripped and ")" in stripped:
        with_parens = _INLINE_PAREN_RE.sub(lambda m: m.group(1), stripped)
        without_parens = _INLINE_PAREN_RE.sub("", stripped)
        # 规范化内部多余空格
        with_parens = re.sub(r"\s+", " ", with_parens).strip()
        without_parens = re.sub(r"\s+", " ", without_parens).strip()
        push(with_parens)
        push(without_parens)

    # 4) 空格/连字符压缩版本
    for v in list(out):
        push(_SPACE_RE.sub("", v))

    return out


def match_word(
    entry: dict, ecdict: Dict[str, dict]
) -> Tuple[Optional[dict], Optional[str], Optional[str]]:
    """
    返回 (ecdict 条目, 实际命中的 key, 命中策略名)；找不到返回 (None, None, None)
    """
    term = _norm(entry.get("term", ""))
    base = _norm(entry.get("baseTerm", ""))
    spellings: List[str] = [_norm(s) for s in entry.get("acceptedSpellings") or []]

    # 1. term
    if term in ecdict:
        return ecdict[term], term, "term"
    # 2. baseTerm
    if base and base != term and base in ecdict:
        return ecdict[base], base, "baseTerm"
    # 3. acceptedSpellings
    for s in spellings:
        if s and s != term and s != base and s in ecdict:
            return ecdict[s], s, "acceptedSpellings"
    # 4. 基于 term 派生的候选词（去括号、去标点、合并空格）
    for cand in _gen_candidates(entry.get("term", "")):
        if cand and cand != term and cand in ecdict:
            return ecdict[cand], cand, "normalized"
    # 5. 短语回退：取首词（仅对认读词尝试）
    head_src = None
    for cand in [term] + _gen_candidates(entry.get("term", "")):
        if " " in cand:
            head_src = cand
            break
    if head_src and entry.get("learningTarget") == "recognize":
        head = head_src.split(" ", 1)[0]
        if head in ecdict:
            return ecdict[head], head, "headWord"
    return None, None, None


# -------------------- 3. enrichment 主流程 --------------------

def enrich() -> None:
    print(f"[1/4] 读取 KET 原始词表: {KET_INPUT}")
    ket: List[dict] = json.loads(KET_INPUT.read_text(encoding="utf-8"))
    print(f"       KET 词条数: {len(ket)}")

    print(f"[2/4] 读取 ECDICT: {ECDICT_CSV}")
    ecdict = load_ecdict(ECDICT_CSV)
    print(f"       ECDICT 词条数: {len(ecdict):,}")

    print("[3/4] 执行 enrichment...")
    enriched: List[dict] = []
    unmatched: List[dict] = []
    strategy_counter: Dict[str, int] = {}

    for entry in ket:
        rec, hit_key, strategy = match_word(entry, ecdict)
        new_entry = dict(entry)
        if rec is not None:
            new_entry["translation"] = rec["translation"]
            new_entry["definition"] = rec["definition"]
            new_entry["phonetic"] = rec["phonetic"]
            new_entry["tag"] = rec["tag"]          # zk/gk/ky/cet4/cet6/toefl/ielts/gre
            new_entry["bncFrq"] = rec["bnc"]       # BNC 词频排名
            new_entry["coca"] = rec["frq"]         # COCA 词频排名
            new_entry["ecdictMatch"] = {
                "key": hit_key,
                "strategy": strategy,
            }
            enriched.append(new_entry)
            strategy_counter[strategy] = strategy_counter.get(strategy, 0) + 1
        else:
            new_entry["translation"] = ""
            new_entry["definition"] = ""
            new_entry["phonetic"] = ""
            new_entry["tag"] = ""
            new_entry["bncFrq"] = ""
            new_entry["coca"] = ""
            new_entry["ecdictMatch"] = None
            enriched.append(new_entry)
            unmatched.append({
                "sourceOrder": entry.get("sourceOrder"),
                "term": entry.get("term"),
                "baseTerm": entry.get("baseTerm"),
                "partOfSpeech": entry.get("partOfSpeech"),
                "learningTarget": entry.get("learningTarget"),
            })

    matched = len(ket) - len(unmatched)
    print(f"       匹配: {matched}/{len(ket)} ({matched / len(ket):.1%})")
    print(f"       未匹配: {len(unmatched)}")
    print(f"       命中策略分布:")
    for s, c in sorted(strategy_counter.items(), key=lambda x: -x[1]):
        print(f"         - {s:<20} {c}")

    print(f"[4/4] 写出结果...")
    OUT_ENRICHED.write_text(
        json.dumps(enriched, ensure_ascii=False, indent=2),
        encoding="utf-8",
    )
    OUT_UNMATCHED.write_text(
        json.dumps(unmatched, ensure_ascii=False, indent=2),
        encoding="utf-8",
    )
    print(f"       -> {OUT_ENRICHED}")
    print(f"       -> {OUT_UNMATCHED}")
    print("Done.")


if __name__ == "__main__":
    enrich()
