package wordbook

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	ErrWordbookNotFound = errors.New("wordbook not found")
	ErrWordbookDisabled = errors.New("wordbook is disabled")
)

// Definition describes one wordbook entry from config.
type Definition struct {
	Code        string `json:"code"`
	ShortName   string `json:"shortName"`
	FullName    string `json:"fullName"`
	CardProfile string `json:"cardProfile,omitempty"`
	JSONPath    string `json:"jsonPath"`
	Importer    string `json:"importer"`
	Enabled     bool   `json:"enabled"`
}

type listPayload struct {
	Wordbooks []Definition `json:"wordbooks"`
}

var (
	mu          sync.RWMutex
	definitions []Definition
	byCode      map[string]Definition
)

// LoadRegistry loads wordbook definitions from a JSON file.
func LoadRegistry(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read wordbook config %s: %w", path, err)
	}

	var payload listPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("parse wordbook config %s: %w", path, err)
	}
	if len(payload.Wordbooks) == 0 {
		return fmt.Errorf("wordbook config %s has no entries", path)
	}

	nextDefs := make([]Definition, 0, len(payload.Wordbooks))
	nextMap := make(map[string]Definition, len(payload.Wordbooks))
	for _, wb := range payload.Wordbooks {
		normalized := normalizeCode(wb.Code)
		if normalized == "" {
			return fmt.Errorf("wordbook code is required")
		}
		if wb.ShortName == "" || wb.FullName == "" {
			return fmt.Errorf("wordbook %s shortName/fullName is required", normalized)
		}
		if wb.JSONPath == "" {
			return fmt.Errorf("wordbook %s jsonPath is required", normalized)
		}
		if wb.Importer == "" {
			return fmt.Errorf("wordbook %s importer is required", normalized)
		}
		if _, exists := nextMap[normalized]; exists {
			return fmt.Errorf("duplicate wordbook code %s", normalized)
		}

		wb.Code = normalized
		nextDefs = append(nextDefs, wb)
		nextMap[normalized] = wb
	}

	mu.Lock()
	definitions = nextDefs
	byCode = nextMap
	mu.Unlock()
	return nil
}

// GetEnabled returns enabled wordbooks in config order.
func GetEnabled() []Definition {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]Definition, 0, len(definitions))
	for _, wb := range definitions {
		if wb.Enabled {
			result = append(result, wb)
		}
	}
	return result
}

// GetAll returns all configured wordbooks in config order.
func GetAll() []Definition {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]Definition, len(definitions))
	copy(result, definitions)
	return result
}

// ValidateEnabled validates whether code exists and is enabled.
func ValidateEnabled(code string) error {
	normalized := normalizeCode(code)

	mu.RLock()
	wb, ok := byCode[normalized]
	mu.RUnlock()

	if !ok {
		return fmt.Errorf("%w: %s", ErrWordbookNotFound, normalized)
	}
	if !wb.Enabled {
		return fmt.Errorf("%w: %s", ErrWordbookDisabled, normalized)
	}
	return nil
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
