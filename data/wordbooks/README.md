# WordTask 词库目录

## 当前词库

| 文件 | 词条数 | 说明 |
| --- | --- | --- |
| `ket-a2-key.json` | 1,721 | KET (A2 Key for Schools) 原始词表，来自 [honorwu/ketwords](https://github.com/honorwu/ketwords)，基于剑桥 2020 官方词表 PDF 解析 |
| `ket-a2-key.enriched.json` | 1,721 | 在原始词表基础上用 ECDICT 补齐中文释义、音标、英文定义、词频标签；**后端/前端请直接使用这份** |

## 字段说明（enriched 版）

```jsonc
{
  // ----- 来自 KET 原始词表 -----
  "sourceOrder": 2,                    // 官方词表中的顺序
  "term": "a few",                     // 显示用词条
  "baseTerm": "a few",                 // 基础形式（用于匹配）
  "normalizedTerm": "a few",           // 全小写、去标点的标准化形式
  "partOfSpeech": "det, adj & pron",   // 词性
  "examples": ["I invited a few of my friends."],
  "acceptedSpellings": ["a few"],      // 所有可接受拼写（英美变体）
  "theme": "general",                  // 主题：general/time/food/travel/school/...
  "priority": "C",                     // 学习优先级：S > A > B > C
  "learningTarget": "recognize",       // recognize | listen | spell
  "spellingRequired": 0,               // 0/1 是否必须拼写

  // ----- 来自 ECDICT enrichment -----
  "translation": "几个, 少数, 一些",      // 中文释义
  "definition": "...",                 // 英文释义（可能为空）
  "phonetic": "ə fjuː",                // IPA 音标（可能为空）
  "tag": "zk gk cet4",                 // 词频标签：zk=中考 gk=高考 ky=考研 cet4/cet6/toefl/ielts/gre
  "bncFrq": "295",                     // BNC 词频排名
  "coca": "385",                       // COCA 词频排名
  "ecdictMatch": {                     // 调试用：命中的 ECDICT key 和策略
    "key": "a few",
    "strategy": "term"
  }
}
```

## 如何复现 / 重新生成 enriched 文件

```powershell
# 1. 下载 ECDICT 基础词典（63 MB，已在 .gitignore 中忽略）
New-Item -ItemType Directory -Force -Path data/ecdict
curl.exe -sSL -o data/ecdict/ecdict.csv `
  https://raw.githubusercontent.com/skywind3000/ECDICT/master/ecdict.csv

# 2. 运行 enrichment 脚本
python scripts/enrich_ket_with_ecdict.py
```

当前脚本匹配率 **100%**（1721/1721）。命中策略分布：

| 策略 | 命中数 | 说明 |
| --- | --- | --- |
| `term` | 1639 | 原始 term 直接精确匹配 |
| `normalized` | 63 | 去掉 `(Am Eng)` 这类标注 / `blond(e)` 这类可选字母后匹配 |
| `acceptedSpellings` | 14 | 英/美拼写变体匹配（color vs colour 等） |
| `headWord` | 5 | 短语取首词兜底（如 `get on` → `get`） |

## 词源与许可

- **KET 原始词表**：剑桥大学考试委员会（Cambridge Assessment English）官方 A2 Key Vocabulary List 2020 版，词表本身处于公共教育用途；引用遵循 Fair Use
- **ECDICT**：[skywind3000/ECDICT](https://github.com/skywind3000/ECDICT) MIT License
- **honorwu/ketwords**：提供了 PDF → JSON 的解析快照
