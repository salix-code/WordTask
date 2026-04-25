# SQLite Schema & Import Mapping

Normative DDL for importing `data/wordbooks/<book>.enriched.json` into the WordTask SQLite database (M3 milestone).

## Tables

```sql
-- One row per shipped wordbook file.
CREATE TABLE IF NOT EXISTS wordbooks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    slug        TEXT    NOT NULL UNIQUE,  -- e.g. "ket-a2-key"
    title       TEXT    NOT NULL,         -- e.g. "KET A2 Key for Schools"
    source_path TEXT    NOT NULL,         -- e.g. "data/wordbooks/ket-a2-key.enriched.json"
    total       INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- One row per entry in a wordbook.
CREATE TABLE IF NOT EXISTS wordbook_entries (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    wordbook_id        INTEGER NOT NULL REFERENCES wordbooks(id) ON DELETE CASCADE,
    source_order       INTEGER NOT NULL,              -- Entry.sourceOrder
    term               TEXT    NOT NULL,              -- Entry.term
    base_term          TEXT    NOT NULL,              -- Entry.baseTerm
    normalized_term    TEXT    NOT NULL,              -- Entry.normalizedTerm (indexed)
    part_of_speech     TEXT    NOT NULL,
    examples_json      TEXT    NOT NULL DEFAULT '[]', -- JSON array of strings
    spellings_json     TEXT    NOT NULL DEFAULT '[]', -- JSON array of strings (acceptedSpellings)
    theme              TEXT    NOT NULL DEFAULT 'general',
    priority           TEXT    NOT NULL CHECK (priority IN ('S','A','B','C')),
    learning_target    TEXT    NOT NULL CHECK (learning_target IN ('recognize','listen','spell')),
    spelling_required  INTEGER NOT NULL CHECK (spelling_required IN (0,1)) DEFAULT 0,

    -- Enrichment (nullable where the source had "")
    translation        TEXT,
    definition         TEXT,
    phonetic           TEXT,
    tag                TEXT,                          -- raw space-separated string
    bnc_frq            INTEGER NOT NULL DEFAULT 0,    -- parsed from string; 0 = unknown
    coca               INTEGER NOT NULL DEFAULT 0,

    UNIQUE (wordbook_id, source_order)
);

CREATE INDEX IF NOT EXISTS idx_entries_normalized
    ON wordbook_entries (normalized_term);

CREATE INDEX IF NOT EXISTS idx_entries_priority
    ON wordbook_entries (wordbook_id, priority);
```

## Field-by-field Import Mapping

| JSON field | Column | Conversion |
| --- | --- | --- |
| `sourceOrder` | `source_order` | direct int |
| `term` | `term` | direct |
| `baseTerm` | `base_term` | direct |
| `normalizedTerm` | `normalized_term` | direct, indexed |
| `partOfSpeech` | `part_of_speech` | direct |
| `examples` | `examples_json` | `json.Marshal` / `JSON.stringify` |
| `acceptedSpellings` | `spellings_json` | same |
| `theme` | `theme` | direct; default `'general'` if empty |
| `priority` | `priority` | direct; CHECK constraint enforces enum |
| `learningTarget` | `learning_target` | direct; CHECK enforces enum |
| `spellingRequired` | `spelling_required` | direct int 0/1 |
| `translation` | `translation` | `""` → `NULL` |
| `definition` | `definition` | `""` → `NULL` |
| `phonetic` | `phonetic` | `""` → `NULL` |
| `tag` | `tag` | `""` → `NULL`; keep raw string |
| `bncFrq` | `bnc_frq` | `strconv.Atoi`; parse errors → `0` |
| `coca` | `coca` | same |
| `ecdictMatch` | — | **drop** (debug-only, not stored) |

## Idempotent Import

Use `INSERT OR REPLACE` keyed on `(wordbook_id, source_order)` so re-running the seed against an updated enriched file overwrites in place without duplicating rows.

```sql
INSERT OR REPLACE INTO wordbook_entries (
    wordbook_id, source_order, term, base_term, normalized_term,
    part_of_speech, examples_json, spellings_json,
    theme, priority, learning_target, spelling_required,
    translation, definition, phonetic, tag, bnc_frq, coca
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
```

## Gotchas

- SQLite stores everything as TEXT if you let it; explicitly pass typed parameters so `bnc_frq` / `coca` / `spelling_required` become INTEGER.
- The source JSON uses curly Unicode quotes (`'`) in some example sentences. Always write/read as UTF-8; do not ASCII-escape.
- `acceptedSpellings` occasionally splits `"a/an"` into three variants — when implementing spelling checks, load the JSON array and compare case-insensitively against any element.
- If a new enrichment source is added later (e.g. Longman definitions), add nullable columns rather than reshaping existing ones; keep the JSON file as the single write-through source of truth.
