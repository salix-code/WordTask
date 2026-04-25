---
name: wordbook-schema
description: This skill should be used whenever CodeBuddy works with WordTask wordbook JSON files under `data/wordbooks/` (including `*.json`, `*.enriched.json`, `*.unmatched.json`). It defines the canonical schema, field semantics, enum values, and ECDICT enrichment contract, so that any code that reads, writes, validates, imports into SQLite, or renders these records in the backend/frontend stays consistent. Load this skill for tasks like designing Go structs, TypeScript interfaces, database import scripts, UI rendering of word cards, adding new wordbooks, or debugging enrichment results.
---

# WordTask Wordbook JSON Schema

## Purpose

WordTask ships curated vocabulary lists (currently KET A2) as JSON under `data/wordbooks/`. Every downstream component — Go backend structs, SQLite import scripts, the Vite+TS frontend word card UI, and the FSRS scheduler — must agree on the same field layout. This skill is the single source of truth for that layout.

## When to Use

Invoke this skill when any of the following apply:

- Reading, writing, or validating files under `data/wordbooks/`.
- Declaring Go structs, TypeScript interfaces, or SQLite table columns that mirror a wordbook entry.
- Writing or updating import/seed scripts that load wordbooks into the database.
- Building UI components that render word cards (showing translation, phonetic, examples, tags).
- Adding a new wordbook (e.g. PET B1, IELTS) or a new enrichment source.
- Debugging mismatches between `*.json` (raw) and `*.enriched.json` (enriched) files.

## File Layout in `data/wordbooks/`

Each wordbook ships as a **top-level JSON array** of entry objects (no wrapper). Three file flavours may coexist per wordbook:

| Suffix | Purpose | Required fields |
| --- | --- | --- |
| `<book>.json` | Raw curated list (KET source) | Core fields only |
| `<book>.enriched.json` | Raw list + ECDICT fields merged in | Core + Enrichment fields |
| `<book>.unmatched.json` | Entries that failed ECDICT lookup (often `[]`) | Core fields only |

**Primary consumption rule**: backend and frontend MUST read `<book>.enriched.json`. The raw `.json` is kept only for provenance and re-running enrichment.

## Entry Object Schema

### Core fields (always present, from the curated source)

| Field | Type | Notes |
| --- | --- | --- |
| `sourceOrder` | `int` | 1-based position in the official source document; stable primary ordering key. |
| `term` | `string` | Display form as printed in the source (may contain slashes like `"a/an"` or parentheticals like `"color (Am Eng)"`). |
| `baseTerm` | `string` | Canonical lemma used for matching; typically equals `term` with source annotations stripped. |
| `normalizedTerm` | `string` | Lowercased, punctuation-stripped form used for fuzzy/search comparisons (e.g. `"a/an"` → `"a an"`). |
| `partOfSpeech` | `string` | POS label; may contain combined values like `"det, adj & pron"` or `"adv & prep"`. Do NOT assume a single token. |
| `examples` | `string[]` | Example sentences/phrases; may be empty `[]`. Entries may contain the POS tag in parentheses, e.g. `"I have about £3. (adv)"`. |
| `acceptedSpellings` | `string[]` | All spellings considered correct for this entry (handles `a/an` → `["a/an","a","an"]` and BrE/AmE variants). Use this array, not `term`, when validating user input. |
| `theme` | `string` (enum) | Topic grouping. See [Enum values](#enum-values). |
| `priority` | `string` (enum) | Learning priority, `S` > `A` > `B` > `C`. |
| `learningTarget` | `string` (enum) | `recognize` \| `listen` \| `spell`. Drives which exercise types are offered. |
| `spellingRequired` | `int` (0 or 1) | Hard flag: if `1`, the user must type the word correctly; treat as boolean. |

### Enrichment fields (present only in `*.enriched.json`)

Injected by `scripts/enrich_ket_with_ecdict.py` from the ECDICT dictionary. All fields are always present in enriched files but may be empty strings when ECDICT has no data.

| Field | Type | Notes |
| --- | --- | --- |
| `translation` | `string` | Chinese gloss, often multi-line (`\n` separated). May include POS prefixes like `"prep. ..."`. |
| `definition` | `string` | English definition (WordNet-style). May be empty `""`. |
| `phonetic` | `string` | IPA without surrounding `/`. May be empty `""`. Example: `"ә'baut"`. |
| `tag` | `string` | Space-separated frequency/exam tags. See [Enum values](#enum-values). May be empty `""`. |
| `bncFrq` | `string` | BNC frequency rank as a **string of digits**; `"0"` means unknown. Parse to int as needed. |
| `coca` | `string` | COCA frequency rank, same string-of-digits convention. |
| `ecdictMatch` | `object` | Debug info: `{ "key": string, "strategy": "term" \| "normalized" \| "acceptedSpellings" \| "headWord" }`. Safe to drop when importing to DB. |

### Enum values

- **`theme`**: `general`, `time`, `food`, `travel`, `school`, ... (open-ended; do not hardcode an exhaustive list — accept any string but default unknown to `general`).
- **`priority`**: `S`, `A`, `B`, `C` (exactly these four, ordered descending).
- **`learningTarget`**: `recognize`, `listen`, `spell`.
- **`ecdictMatch.strategy`**: `term`, `normalized`, `acceptedSpellings`, `headWord`.
- **`tag`** tokens (space-separated inside the string): `zk` (中考), `gk` (高考), `ky` (考研), `cet4`, `cet6`, `toefl`, `ielts`, `gre`. Split on whitespace; treat empty string as "no tags".

## Canonical Example

```json
{
  "sourceOrder": 2,
  "term": "a few",
  "baseTerm": "a few",
  "normalizedTerm": "a few",
  "partOfSpeech": "det, adj & pron",
  "examples": [
    "I invited a few of my friends.",
    "I'll be ready in a few minutes."
  ],
  "acceptedSpellings": ["a few"],
  "theme": "general",
  "priority": "C",
  "learningTarget": "recognize",
  "spellingRequired": 0,
  "translation": "几个, 少数, 一些",
  "definition": "",
  "phonetic": "",
  "tag": "",
  "bncFrq": "0",
  "coca": "0",
  "ecdictMatch": { "key": "a few", "strategy": "term" }
}
```

## How to Use This Skill

### When designing Go / TypeScript types

Mirror the core schema exactly (field names in camelCase). Key reminders:

- `sourceOrder`, `spellingRequired` are integers; everything in the enrichment block except `ecdictMatch` is a **string** (including `bncFrq` and `coca` — do NOT type them as int).
- `examples` and `acceptedSpellings` default to `[]`, never `null`.
- `priority` / `learningTarget` should be modeled as enums/string literal unions to catch typos.
- `ecdictMatch` can be omitted from public API DTOs; keep it only in import/debug paths.

Ready-to-use snippets live in [`references/type-definitions.md`](references/type-definitions.md).

### When importing into SQLite

- Use `sourceOrder` + wordbook id as the composite natural key; add a surrogate `id INTEGER PRIMARY KEY` for FK convenience.
- Store `examples`, `acceptedSpellings` either as JSON text columns or in a normalized child table. Prefer JSON text for MVP (matches the current M3 roadmap).
- Parse `bncFrq` / `coca` to `INTEGER` on write; store `0` when the source is `"0"` or empty.
- Split `tag` on whitespace into a tag set if tag filtering is needed.

A concrete DDL template is in [`references/sqlite-schema.md`](references/sqlite-schema.md).

### When rendering word cards in the frontend

- Title: `term`; fallback pronunciation hint: `phonetic` (wrap with `/ /` when displaying).
- Chinese meaning: `translation` — render with `white-space: pre-line` because it may contain `\n`.
- Examples: iterate `examples`; strip trailing ` (adv)` / ` (prep)` parentheticals when showing to learners, but preserve for POS-specific views.
- Use `acceptedSpellings` (not `term`) to judge user input in spelling exercises.
- Use Web Speech API with `term` (or `acceptedSpellings[0]`) as the utterance text.

### When adding a new wordbook

1. Drop the raw `<book>.json` array into `data/wordbooks/`, matching the core schema above.
2. Run `python scripts/enrich_ket_with_ecdict.py` (or a generalized variant) to produce `<book>.enriched.json` and `<book>.unmatched.json`.
3. Update `data/wordbooks/README.md` with the new row and current match rate.
4. Verify the file is a **top-level JSON array** and every entry has all core fields.

### When debugging enrichment

- `ecdictMatch.strategy` tells why a row matched: `term` (exact), `normalized` (after stripping `(Am Eng)` / optional letters like `blond(e)`), `acceptedSpellings` (variant hit), `headWord` (first token of a phrase, e.g. `get on` → `get`).
- Empty `translation`/`definition`/`phonetic` with non-empty `ecdictMatch` = ECDICT row was found but lacked that column. Not a bug.
- A row appears in `*.unmatched.json` only when all four strategies fail; the file is typically `[]` (current KET match rate is 100%).

## References

- [`references/type-definitions.md`](references/type-definitions.md) — Ready-to-copy Go structs, TypeScript interfaces, and JSON Schema.
- [`references/sqlite-schema.md`](references/sqlite-schema.md) — SQLite DDL and import mapping guidance.
- Upstream doc: `data/wordbooks/README.md` (keep in sync; this skill is the normative reference for code, the README is user-facing documentation).
