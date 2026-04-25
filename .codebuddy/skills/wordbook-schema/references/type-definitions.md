# Type Definitions

Ready-to-use type declarations mirroring the wordbook entry schema. Keep these in sync with `SKILL.md` — if you change one, change the other.

## Go (backend)

```go
package wordbook

// Entry represents one vocabulary entry as stored in
// data/wordbooks/<book>.enriched.json.
//
// Field order and JSON tags MUST match the on-disk layout; omitempty is
// intentionally NOT used so round-tripping preserves explicit empty values.
type Entry struct {
    // Core fields (always present)
    SourceOrder       int      `json:"sourceOrder"`
    Term              string   `json:"term"`
    BaseTerm          string   `json:"baseTerm"`
    NormalizedTerm    string   `json:"normalizedTerm"`
    PartOfSpeech      string   `json:"partOfSpeech"`
    Examples          []string `json:"examples"`
    AcceptedSpellings []string `json:"acceptedSpellings"`
    Theme             string   `json:"theme"`
    Priority          Priority `json:"priority"`         // "S" | "A" | "B" | "C"
    LearningTarget    Target   `json:"learningTarget"`   // "recognize" | "listen" | "spell"
    SpellingRequired  int      `json:"spellingRequired"` // 0 or 1

    // Enrichment fields (present in *.enriched.json; may be empty strings)
    Translation  string        `json:"translation"`
    Definition   string        `json:"definition"`
    Phonetic     string        `json:"phonetic"`
    Tag          string        `json:"tag"`    // space-separated: "zk gk cet4"
    BNCFrq       string        `json:"bncFrq"` // digit string; "0" = unknown
    COCA         string        `json:"coca"`   // digit string; "0" = unknown
    EcdictMatch  *EcdictMatch  `json:"ecdictMatch,omitempty"`
}

type Priority string

const (
    PriorityS Priority = "S"
    PriorityA Priority = "A"
    PriorityB Priority = "B"
    PriorityC Priority = "C"
)

type Target string

const (
    TargetRecognize Target = "recognize"
    TargetListen    Target = "listen"
    TargetSpell     Target = "spell"
)

type EcdictMatch struct {
    Key      string `json:"key"`
    Strategy string `json:"strategy"` // term | normalized | acceptedSpellings | headWord
}

// MustSpell returns true when the entry requires the learner to type the word.
func (e *Entry) MustSpell() bool { return e.SpellingRequired == 1 }

// Tags splits the space-separated Tag field.
func (e *Entry) Tags() []string {
    if e.Tag == "" {
        return nil
    }
    return strings.Fields(e.Tag) // import "strings"
}
```

## TypeScript (frontend)

```ts
// src/types/wordbook.ts

export type Priority = "S" | "A" | "B" | "C";
export type LearningTarget = "recognize" | "listen" | "spell";
export type EcdictStrategy =
  | "term"
  | "normalized"
  | "acceptedSpellings"
  | "headWord";

export interface EcdictMatch {
  key: string;
  strategy: EcdictStrategy;
}

/** One entry as shipped in data/wordbooks/<book>.enriched.json. */
export interface WordbookEntry {
  // Core
  sourceOrder: number;
  term: string;
  baseTerm: string;
  normalizedTerm: string;
  partOfSpeech: string;
  examples: string[];
  acceptedSpellings: string[];
  theme: string;
  priority: Priority;
  learningTarget: LearningTarget;
  /** 0 or 1; use Boolean(spellingRequired) where a bool is needed. */
  spellingRequired: 0 | 1;

  // Enrichment (may be empty strings)
  translation: string;
  definition: string;
  phonetic: string;
  /** Space-separated tags: "zk gk cet4". Split on whitespace. */
  tag: string;
  /** Digit string; "0" means unknown. */
  bncFrq: string;
  /** Digit string; "0" means unknown. */
  coca: string;
  ecdictMatch?: EcdictMatch;
}

/** A whole wordbook file is a top-level array. */
export type Wordbook = WordbookEntry[];

// Helpers ---------------------------------------------------------------

export const mustSpell = (e: WordbookEntry): boolean =>
  e.spellingRequired === 1;

export const tagList = (e: WordbookEntry): string[] =>
  e.tag ? e.tag.split(/\s+/).filter(Boolean) : [];

export const freqBNC = (e: WordbookEntry): number | null => {
  const n = parseInt(e.bncFrq, 10);
  return Number.isFinite(n) && n > 0 ? n : null;
};
```

## JSON Schema (for CI validation)

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "WordTask Wordbook",
  "type": "array",
  "items": {
    "type": "object",
    "required": [
      "sourceOrder", "term", "baseTerm", "normalizedTerm",
      "partOfSpeech", "examples", "acceptedSpellings",
      "theme", "priority", "learningTarget", "spellingRequired"
    ],
    "properties": {
      "sourceOrder":       { "type": "integer", "minimum": 1 },
      "term":              { "type": "string", "minLength": 1 },
      "baseTerm":          { "type": "string", "minLength": 1 },
      "normalizedTerm":    { "type": "string", "minLength": 1 },
      "partOfSpeech":      { "type": "string" },
      "examples":          { "type": "array", "items": { "type": "string" } },
      "acceptedSpellings": { "type": "array", "items": { "type": "string" }, "minItems": 1 },
      "theme":             { "type": "string" },
      "priority":          { "enum": ["S", "A", "B", "C"] },
      "learningTarget":    { "enum": ["recognize", "listen", "spell"] },
      "spellingRequired":  { "type": "integer", "enum": [0, 1] },

      "translation": { "type": "string" },
      "definition":  { "type": "string" },
      "phonetic":    { "type": "string" },
      "tag":         { "type": "string" },
      "bncFrq":      { "type": "string", "pattern": "^[0-9]+$" },
      "coca":        { "type": "string", "pattern": "^[0-9]+$" },
      "ecdictMatch": {
        "type": "object",
        "required": ["key", "strategy"],
        "properties": {
          "key":      { "type": "string" },
          "strategy": { "enum": ["term", "normalized", "acceptedSpellings", "headWord"] }
        }
      }
    },
    "additionalProperties": false
  }
}
```
