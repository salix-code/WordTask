package api

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

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
	UserID   string `json:"userId" binding:"required"`
	Wordbook string `json:"wordbook" binding:"required"`
	WordID   uint   `json:"wordId" binding:"required"`
	Quality  string `json:"quality" binding:"required,oneof=forgot vague known"`
}

// SubmitReview handles POST /api/words/review.
// When quality is 'known', the word is marked as known in the active cycle.
func SubmitReview(c *gin.Context) {
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "invalid params: "+err.Error())
		return
	}

	if req.Quality == "known" {
		if err := service.MarkWordKnown(req.UserID, req.Wordbook, req.WordID); err != nil {
			log.Printf("[SubmitReview] MarkWordKnown error: %v", err)
			Fail(c, 500, "internal server error")
			return
		}
	}

	OK(c, gin.H{
		"word_id": req.WordID,
		"quality": req.Quality,
	})
}

// AdvanceProgressRequest is kept for backwards compatibility.
type AdvanceProgressRequest struct {
	UserID   string `json:"userId"`
	Wordbook string `json:"wordbook"`
	Count    int    `json:"count"`
}

// AdvanceProgress is a no-op kept for backwards compatibility.
// Per-word progress is now handled by SubmitReview (quality=known).
func AdvanceProgress(c *gin.Context) {
	OK(c, gin.H{"ok": true})
}

// SetupCycleRequest is the request body for POST /api/cycles/setup.
type SetupCycleRequest struct {
	UserID   string   `json:"userId"   binding:"required"`
	Wordbook string   `json:"wordbook" binding:"required"`
	Words    []string `json:"words"    binding:"required"`
}

// SetupCycle handles POST /api/cycles/setup.
// It validates the submitted word list and, if all words pass, creates a new cycle.
func SetupCycle(c *gin.Context) {
	var req SetupCycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "invalid params: "+err.Error())
		return
	}

	result, err := service.ValidateAndCreateCycle(req.UserID, req.Wordbook, req.Words)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCycleInput) {
			Fail(c, 400, err.Error())
			return
		}
		log.Printf("[SetupCycle] error: %v", err)
		Fail(c, 500, "internal server error")
		return
	}

	OK(c, result)
}

// LoginRequest is the request body for POST /api/account/login.
type LoginRequest struct {
	AccountName string `json:"accountName" binding:"required"`
}

// Login handles POST /api/account/login.
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "invalid request body")
		return
	}

	accountService := service.AccountService{}
	account, err := accountService.FindByName(req.AccountName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Fail(c, 404, "account not found")
			return
		}
		log.Printf("[Login] error: %v", err)
		Fail(c, 500, "internal server error")
		return
	}

	OK(c, gin.H{
		"userId":      account.ID,
		"accountName": account.Name,
	})
}
