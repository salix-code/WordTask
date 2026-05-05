package service

import (
	"errors"
	"math"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wordtask-server/internal/db"
	"wordtask-server/internal/model"
	wb "wordtask-server/internal/wordbook"
)

const defaultEaseFactor = 2.5
const minEaseFactor = 1.3

// ReviewDueResult is the response body for GET /api/words/review/due.
type ReviewDueResult struct {
	Words []WordResponse `json:"words"`
}

// GetDueReviewWords returns due words from completed cycles using SM-2 schedule.
func GetDueReviewWords(userID, wordbook string, limit int) (*ReviewDueResult, error) {
	if err := wb.ValidateEnabled(wordbook); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = DailyBatchSize
	}

	if err := ensureReviewProgressRows(userID, wordbook); err != nil {
		return nil, err
	}

	now := time.Now()
	var progressRows []model.ReviewProgress
	if err := db.DB.
		Where("user_id = ? AND wordbook = ? AND due_at <= ?", userID, wordbook, now).
		Order("due_at ASC").
		Limit(limit).
		Find(&progressRows).Error; err != nil {
		return nil, err
	}

	if len(progressRows) == 0 {
		return &ReviewDueResult{Words: []WordResponse{}}, nil
	}

	wordIDs := make([]uint, 0, len(progressRows))
	for _, p := range progressRows {
		wordIDs = append(wordIDs, p.WordID)
	}

	var words []model.Word
	if err := db.DB.
		Where("wordbook = ? AND id IN ?", wordbook, wordIDs).
		Find(&words).Error; err != nil {
		return nil, err
	}

	wordMap := make(map[uint]model.Word, len(words))
	for _, w := range words {
		wordMap[w.ID] = w
	}

	result := make([]WordResponse, 0, len(progressRows))
	for _, p := range progressRows {
		w, ok := wordMap[p.WordID]
		if !ok {
			continue
		}
		result = append(result, toWordResponse(w))
	}
	return &ReviewDueResult{Words: result}, nil
}

// SubmitRevision applies SM-2 update for one review result.
func SubmitRevision(userID, wordbook string, wordID uint, quality string) error {
	if err := wb.ValidateEnabled(wordbook); err != nil {
		return err
	}

	now := time.Now()
	var p model.ReviewProgress
	err := db.DB.
		Where("user_id = ? AND wordbook = ? AND word_id = ?", userID, wordbook, wordID).
		First(&p).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		// Seed on first touch (e.g. old data before review_progress table existed).
		p = model.ReviewProgress{
			UserID:       userID,
			Wordbook:     wordbook,
			WordID:       wordID,
			Repetition:   0,
			IntervalDays: 0,
			EaseFactor:   defaultEaseFactor,
			DueAt:        now,
		}
		if err2 := db.DB.Create(&p).Error; err2 != nil {
			return err2
		}
	}

	q := sm2QualityScore(quality)
	nextInterval, nextRep, nextEF := sm2Next(p.IntervalDays, p.Repetition, p.EaseFactor, q)
	nextDue := now.AddDate(0, 0, nextInterval)

	return db.DB.Model(&model.ReviewProgress{}).
		Where("user_id = ? AND wordbook = ? AND word_id = ?", userID, wordbook, wordID).
		Updates(map[string]interface{}{
			"interval_days":    nextInterval,
			"repetition":       nextRep,
			"ease_factor":      nextEF,
			"due_at":           nextDue,
			"last_reviewed_at": now,
		}).Error
}

func ensureReviewProgressRows(userID, wordbook string) error {
	var wordIDs []uint
	if err := db.DB.
		Table("word_cycles").
		Distinct("word_cycles.word_id").
		Joins("JOIN cycles ON cycles.id = word_cycles.cycle_id").
		Where("cycles.user_id = ? AND cycles.wordbook = ? AND cycles.status = ? AND word_cycles.status = ?", userID, wordbook, "completed", "known").
		Pluck("word_cycles.word_id", &wordIDs).Error; err != nil {
		return err
	}
	if len(wordIDs) == 0 {
		return nil
	}

	now := time.Now()
	rows := make([]model.ReviewProgress, 0, len(wordIDs))
	for _, id := range wordIDs {
		rows = append(rows, model.ReviewProgress{
			UserID:       userID,
			Wordbook:     wordbook,
			WordID:       id,
			Repetition:   0,
			IntervalDays: 0,
			EaseFactor:   defaultEaseFactor,
			DueAt:        now,
		})
	}

	return db.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

func sm2QualityScore(quality string) int {
	switch quality {
	case "known":
		return 5
	case "vague":
		return 3
	default:
		return 1
	}
}

func sm2Next(intervalDays, repetition int, easeFactor float64, quality int) (int, int, float64) {
	if easeFactor < minEaseFactor {
		easeFactor = defaultEaseFactor
	}

	newEF := easeFactor + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
	if newEF < minEaseFactor {
		newEF = minEaseFactor
	}

	if quality < 3 {
		return 1, 0, newEF
	}

	if repetition == 0 {
		return 1, 1, newEF
	}
	if repetition == 1 {
		return 6, 2, newEF
	}

	nextInterval := int(math.Round(float64(intervalDays) * newEF))
	if nextInterval < 1 {
		nextInterval = 1
	}
	return nextInterval, repetition + 1, newEF
}
