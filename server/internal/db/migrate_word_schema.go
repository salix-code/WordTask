package db

import (
	"fmt"

	"wordtask-server/internal/model"
)

func EnsureWordSchema() error {
	hasWords, err := tableExists("words")
	if err != nil {
		return err
	}
	hasWordKET, err := tableExists("word_ket")
	if err != nil {
		return err
	}
	if hasWords && !hasWordKET {
		if err := DB.Exec("ALTER TABLE words RENAME TO word_ket").Error; err != nil {
			return err
		}
	} else if hasWords && hasWordKET {
		var wordKETCount int64
		if err := DB.Raw("SELECT COUNT(1) FROM word_ket").Scan(&wordKETCount).Error; err != nil {
			return err
		}
		if wordKETCount == 0 {
			if err := DB.Exec(`
				INSERT INTO word_ket
				(id, wordbook, source_order, term, phonetic, part_of_speech, translation, examples, theme, priority, learning_target, cycle)
				SELECT id, wordbook, source_order, term, phonetic, part_of_speech, translation, examples, theme, priority, learning_target, 0
				FROM words
			`).Error; err != nil {
				return err
			}
		}
		if err := DB.Exec("DROP TABLE IF EXISTS words").Error; err != nil {
			return err
		}
	}

	hasWordKET, err = tableExists("word_ket")
	if err != nil {
		return err
	}
	if hasWordKET {
		hasCycle, err := columnExists("word_ket", "cycle")
		if err != nil {
			return err
		}
		if !hasCycle {
			if err := DB.Exec("ALTER TABLE word_ket ADD COLUMN cycle INTEGER NOT NULL DEFAULT 0").Error; err != nil {
				return err
			}
		}
	}

	hasCycleNo, err := columnExists("cycles", "cycle_no")
	if err != nil {
		return err
	}
	if !hasCycleNo {
		if err := DB.Exec("ALTER TABLE cycles ADD COLUMN cycle_no INTEGER NOT NULL DEFAULT 0").Error; err != nil {
			return err
		}
	}
	if err := DB.Exec("UPDATE cycles SET cycle_no = id WHERE cycle_no = 0").Error; err != nil {
		return err
	}

	hasWordCycles, err := tableExists("word_cycles")
	if err != nil {
		return err
	}
	if hasWordCycles && hasWordKET {
		if err := DB.Exec(`
			UPDATE word_ket
			SET cycle = (
				SELECT wc.cycle_id
				FROM word_cycles wc
				JOIN cycles c ON c.id = wc.cycle_id
				WHERE wc.word_id = word_ket.id
				  AND c.wordbook = word_ket.wordbook
				  AND c.user_id = ?
				ORDER BY wc.cycle_id ASC
				LIMIT 1
			)
			WHERE cycle = 0
		`, model.AdminUserID).Error; err != nil {
			return err
		}

		if err := DB.Exec(`
			UPDATE word_ket
			SET cycle = (
				SELECT wc.cycle_id
				FROM word_cycles wc
				WHERE wc.word_id = word_ket.id
				ORDER BY wc.cycle_id ASC
				LIMIT 1
			)
			WHERE cycle = 0
		`).Error; err != nil {
			return err
		}

		if err := DB.Exec("DROP TABLE IF EXISTS word_cycles").Error; err != nil {
			return err
		}
	}

	return nil
}

func tableExists(name string) (bool, error) {
	var cnt int64
	if err := DB.Raw("SELECT COUNT(1) FROM sqlite_master WHERE type = 'table' AND name = ?", name).Scan(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func columnExists(tableName, columnName string) (bool, error) {
	type col struct {
		Name string `gorm:"column:name"`
	}
	var cols []col
	if err := DB.Raw(fmt.Sprintf("PRAGMA table_info(%s)", tableName)).Scan(&cols).Error; err != nil {
		return false, err
	}
	for _, c := range cols {
		if c.Name == columnName {
			return true, nil
		}
	}
	return false, nil
}
