package service

import (
	"encoding/json"

	"wordtask-server/internal/db"
	"wordtask-server/internal/model"
)

// DailyGoal is the number of words returned per session, configured here for now.
const DailyGoal = 20

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
	Completed bool           `json:"completed"`
	Words     []WordResponse `json:"words"`
}

// GetTodayWords returns the next DailyGoal words for the user based on their progress offset.
// Progress is never reset automatically; it simply continues from where the user left off.
func GetTodayWords(userID, wordbook string) (*TodayWordsResult, error) {
	// Load or create progress record.
	var progress model.UserProgress
	result := db.DB.Where("user_id = ? AND wordbook = ?", userID, wordbook).First(&progress)
	if result.Error != nil {
		progress = model.UserProgress{UserID: userID, Wordbook: wordbook, CompletedCount: 0}
		if err := db.DB.Create(&progress).Error; err != nil {
			return nil, err
		}
	}

	// Check total words available for this wordbook.
	var totalCount int64
	db.DB.Model(&model.Word{}).Where("wordbook = ?", wordbook).Count(&totalCount)

	if int64(progress.CompletedCount) >= totalCount {
		return &TodayWordsResult{
			DailyGoal: DailyGoal,
			Completed: true,
			Words:     []WordResponse{},
		}, nil
	}

	// Fetch the next DailyGoal words starting from the user's offset.
	var words []model.Word
	if err := db.DB.
		Where("wordbook = ?", wordbook).
		Order("source_order ASC").
		Limit(DailyGoal).
		Offset(progress.CompletedCount).
		Find(&words).Error; err != nil {
		return nil, err
	}

	responses := make([]WordResponse, 0, len(words))
	for _, w := range words {
		responses = append(responses, toWordResponse(w))
	}

	return &TodayWordsResult{
		DailyGoal: DailyGoal,
		Completed: false,
		Words:     responses,
	}, nil
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
