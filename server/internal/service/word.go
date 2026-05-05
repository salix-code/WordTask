package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"wordtask-server/internal/db"
	"wordtask-server/internal/model"
)

// CycleSize is the default/reference number of words in one learning cycle.
// Actual cycle size is flexible and determined by the submitted word list.
const CycleSize = 30

// DailyBatchSize is the maximum number of new words shown per calendar day within a cycle.
const DailyBatchSize = 5

// DefinitionItem maps to the frontend definitions field.
type DefinitionItem struct {
	Pos     string `json:"pos,omitempty"`
	Meaning string `json:"meaning"`
}

// ExampleItem maps to the frontend examples field.
type ExampleItem struct {
	En string `json:"en"`
}

// WordResponse is the DTO sent to the frontend.
type WordResponse struct {
	ID          uint             `json:"id"`
	SortKey     int              `json:"sortKey"`
	Text        string           `json:"text"`
	Phonetic    string           `json:"phonetic,omitempty"`
	Definitions []DefinitionItem `json:"definitions"`
	Examples    []ExampleItem    `json:"examples,omitempty"`
}

// TodayWordsResult is the full response body for GET /api/words/today.
type TodayWordsResult struct {
	DailyGoal int            `json:"dailyGoal"`
	NoCycle   bool           `json:"noCycle,omitempty"`
	Completed bool           `json:"completed"`
	Words     []WordResponse `json:"words"`
}

// GetTodayWords returns up to DailyBatchSize words for the user's active cycle.
// If no ongoing cycle exists but queued cycles exist, the oldest queued cycle
// is promoted to ongoing automatically.
// The daily batch advances once per calendar day (midnight boundary).
func GetTodayWords(userID, wordbook string) (*TodayWordsResult, error) {
	if err := ensureUserCyclesFromAdmin(userID, wordbook); err != nil {
		return nil, err
	}
	wordTable, err := db.ResolveWordTableName(wordbook)
	if err != nil {
		return nil, err
	}

	// 1. Load or create UserProgress.
	var progress model.UserProgress
	if err := db.DB.
		Where("user_id = ? AND wordbook = ?", userID, wordbook).
		FirstOrCreate(&progress, model.UserProgress{UserID: userID, Wordbook: wordbook}).Error; err != nil {
		return nil, err
	}

	// 2. Find the active (ongoing) cycle.
	var cycle model.Cycle
	err = db.DB.
		Where("user_id = ? AND wordbook = ? AND status = ?", userID, wordbook, "ongoing").
		First(&cycle).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// No active cycle — try promote oldest queued cycle.
		var next model.Cycle
		nextErr := db.DB.
			Where("user_id = ? AND wordbook = ? AND status = ?", userID, wordbook, "queued").
			Order("created_at ASC, id ASC").
			First(&next).Error
		if nextErr != nil && !errors.Is(nextErr, gorm.ErrRecordNotFound) {
			return nil, nextErr
		}
		if errors.Is(nextErr, gorm.ErrRecordNotFound) {
			// No active and no queued cycle — ask user to set one up.
			return &TodayWordsResult{
				DailyGoal: DailyBatchSize,
				NoCycle:   true,
				Completed: false,
				Words:     []WordResponse{},
			}, nil
		}
		if err := db.DB.Model(&next).Update("status", "ongoing").Error; err != nil {
			return nil, err
		}
		cycle = next
	}

	// 3. Advance the daily batch pointer on a new calendar day.
	today := truncateToDay(time.Now())
	lastDate := truncateToDay(progress.LastSessionDate)
	if lastDate.Before(today) {
		var knownCount int64
		db.DB.
			Model(&model.ReviewProgress{}).
			Distinct("review_progresses.word_id").
			Joins(fmt.Sprintf("JOIN %s ON %s.id = review_progresses.word_id", wordTable, wordTable)).
			Where(fmt.Sprintf("review_progresses.user_id = ? AND review_progresses.wordbook = ? AND %s.wordbook = ? AND %s.cycle = ?", wordTable, wordTable), userID, wordbook, wordbook, cycle.CycleNo).
			Count(&knownCount)
		progress.LastSessionCompleted = int(knownCount)
		progress.LastSessionDate = time.Now()
		db.DB.Model(&progress).Updates(map[string]interface{}{
			"last_session_date":      progress.LastSessionDate,
			"last_session_completed": progress.LastSessionCompleted,
		})
	}

	// 4. Return up to DailyBatchSize unlearned words from the current cycle.
	var rows []model.Word
	if err2 := db.DB.
		Table(wordTable).
		Where(fmt.Sprintf("%s.wordbook = ? AND %s.cycle = ?", wordTable, wordTable), wordbook, cycle.CycleNo).
		Where(fmt.Sprintf("%s.id NOT IN (?)", wordTable),
			db.DB.Model(&model.ReviewProgress{}).
				Select("review_progresses.word_id").
				Where("review_progresses.user_id = ? AND review_progresses.wordbook = ?", userID, wordbook),
		).
		Order(fmt.Sprintf("%s.source_order ASC", wordTable)).
		Limit(DailyBatchSize).
		Scan(&rows).Error; err2 != nil {
		return nil, err2
	}

	responses := make([]WordResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, toWordResponse(row))
	}

	return &TodayWordsResult{
		DailyGoal: DailyBatchSize,
		Completed: false,
		Words:     responses,
	}, nil
}

func truncateToDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func toWordResponse(w model.Word) WordResponse {
	var rawExamples []string
	_ = json.Unmarshal([]byte(w.Examples), &rawExamples)

	examples := make([]ExampleItem, 0, len(rawExamples))
	for _, e := range rawExamples {
		examples = append(examples, ExampleItem{En: e})
	}

	return WordResponse{
		ID:       w.ID,
		SortKey:  w.SourceOrder,
		Text:     w.Term,
		Phonetic: w.Phonetic,
		Definitions: []DefinitionItem{
			{Pos: w.PartOfSpeech, Meaning: w.Translation},
		},
		Examples: examples,
	}
}
