package api

import (
	"log"

	"github.com/gin-gonic/gin"

	"wordtask-server/internal/service"
)

// Ping handles health check.
func Ping(c *gin.Context) {
	OK(c, gin.H{"message": "pong"})
}

// GetTodayWords handles GET /api/words/today?userId=xxx&wordbook=KET
func GetTodayWords(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		Fail(c, 400, "userId is required")
		return
	}
	wordbook := c.Query("wordbook")
	if wordbook == "" {
		Fail(c, 400, "wordbook is required")
		return
	}

	result, err := service.GetTodayWords(userID, wordbook)
	if err != nil {
		log.Printf("[GetTodayWords] error: %v", err)
		Fail(c, 500, "internal server error")
		return
	}

	OK(c, result)
}

// ReviewRequest is the request body for POST /api/words/review.
type ReviewRequest struct {
	WordID int    `json:"word_id" binding:"required"`
	Grade  string `json:"grade" binding:"required,oneof=red yellow green"`
}

// SubmitReview handles POST /api/words/review.
// Grade persistence will be wired to FSRS in M4.
func SubmitReview(c *gin.Context) {
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "invalid params: "+err.Error())
		return
	}

	log.Printf("[Review] word_id=%d, grade=%s", req.WordID, req.Grade)

	OK(c, gin.H{
		"word_id":        req.WordID,
		"grade":          req.Grade,
		"next_review_at": nil,
	})
}

// AdvanceProgressRequest is the request body for POST /api/progress/advance.
type AdvanceProgressRequest struct {
	UserID   string `json:"userId" binding:"required"`
	Wordbook string `json:"wordbook" binding:"required"`
	Count    int    `json:"count" binding:"required,min=1"`
}

// AdvanceProgress handles POST /api/progress/advance.
func AdvanceProgress(c *gin.Context) {
	var req AdvanceProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "invalid params: "+err.Error())
		return
	}

	if err := service.AdvanceProgress(req.UserID, req.Wordbook, req.Count); err != nil {
		log.Printf("[AdvanceProgress] error: %v", err)
		Fail(c, 500, "internal server error")
		return
	}

	OK(c, gin.H{"ok": true})
}
