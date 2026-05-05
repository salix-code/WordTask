package api

import (
	"errors"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"wordtask-server/internal/service"
	"wordtask-server/internal/wordbook"
)

// Ping handles health check.
func Ping(c *gin.Context) {
	OK(c, gin.H{"message": "pong"})
}

// ListWordbooks handles GET /api/wordbooks.
func ListWordbooks(c *gin.Context) {
	enabled := wordbook.GetEnabled()
	items := make([]gin.H, 0, len(enabled))
	for _, wb := range enabled {
		items = append(items, gin.H{
			"code":      wb.Code,
			"shortName": wb.ShortName,
			"fullName":  wb.FullName,
		})
	}
	OK(c, gin.H{"wordbooks": items})
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
		if isWordbookRequestError(err) {
			Fail(c, 400, err.Error())
			return
		}
		log.Printf("[GetTodayWords] error: %v", err)
		Fail(c, 500, "internal server error")
		return
	}

	OK(c, result)
}

// GetReviewDueWords handles GET /api/words/review/due?userId=xxx&wordbook=KET&limit=5
func GetReviewDueWords(c *gin.Context) {
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

	limit := 0
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			Fail(c, 400, "limit must be a positive integer")
			return
		}
		limit = parsed
	}

	result, err := service.GetDueReviewWords(userID, wordbook, limit)
	if err != nil {
		if isWordbookRequestError(err) {
			Fail(c, 400, err.Error())
			return
		}
		log.Printf("[GetReviewDueWords] error: %v", err)
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
			if isWordbookRequestError(err) {
				Fail(c, 400, err.Error())
				return
			}
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

// SubmitRevision handles POST /api/words/review/revision.
func SubmitRevision(c *gin.Context) {
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "invalid params: "+err.Error())
		return
	}

	if err := service.SubmitRevision(req.UserID, req.Wordbook, req.WordID, req.Quality); err != nil {
		if isWordbookRequestError(err) {
			Fail(c, 400, err.Error())
			return
		}
		log.Printf("[SubmitRevision] error: %v", err)
		Fail(c, 500, "internal server error")
		return
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
	if !requireAdminUser(c, req.UserID) {
		return
	}

	result, err := service.ValidateAndCreateCycle(req.UserID, req.Wordbook, req.Words)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCycleInput) {
			Fail(c, 400, err.Error())
			return
		}
		if isWordbookRequestError(err) {
			Fail(c, 400, err.Error())
			return
		}
		log.Printf("[SetupCycle] error: %v", err)
		Fail(c, 500, "internal server error")
		return
	}

	OK(c, result)
}

// GetCurrentCycle handles GET /api/cycles/current?userId=xxx&wordbook=KET
func GetCurrentCycle(c *gin.Context) {
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

	words, err := service.GetCurrentCycleWords(userID, wordbook)
	if err != nil {
		if isWordbookRequestError(err) {
			Fail(c, 400, err.Error())
			return
		}
		log.Printf("[GetCurrentCycle] error: %v", err)
		Fail(c, 500, "internal server error")
		return
	}

	if words == nil {
		OK(c, gin.H{"hasCycle": false, "words": []interface{}{}})
		return
	}
	OK(c, gin.H{"hasCycle": true, "words": words})
}

// ListCycles handles GET /api/cycles?userId=xxx&wordbook=KET
func ListCycles(c *gin.Context) {
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

	cycles, err := service.ListCycles(userID, wordbook)
	if err != nil {
		if isWordbookRequestError(err) {
			Fail(c, 400, err.Error())
			return
		}
		log.Printf("[ListCycles] error: %v", err)
		Fail(c, 500, "internal server error")
		return
	}

	OK(c, gin.H{"cycles": cycles})
}

// GetCycleDetail handles GET /api/cycles/:id?userId=xxx&wordbook=KET
func GetCycleDetail(c *gin.Context) {
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

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		Fail(c, 400, "invalid cycle id")
		return
	}

	detail, err := service.GetCycleWordsByID(userID, wordbook, uint(id))
	if err != nil {
		if errors.Is(err, service.ErrInvalidCycleInput) {
			Fail(c, 400, err.Error())
			return
		}
		if isWordbookRequestError(err) {
			Fail(c, 400, err.Error())
			return
		}
		log.Printf("[GetCycleDetail] error: %v", err)
		Fail(c, 500, "internal server error")
		return
	}

	OK(c, detail)
}

// UpdateCycleRequest is the request body for PUT /api/cycles/current.
type UpdateCycleRequest struct {
	UserID   string   `json:"userId"   binding:"required"`
	Wordbook string   `json:"wordbook" binding:"required"`
	Words    []string `json:"words"    binding:"required"`
}

// UpdateCycle handles PUT /api/cycles/current.
func UpdateCycle(c *gin.Context) {
	var req UpdateCycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "invalid params: "+err.Error())
		return
	}
	if !requireAdminUser(c, req.UserID) {
		return
	}

	result, err := service.UpdateCurrentCycle(req.UserID, req.Wordbook, req.Words)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCycleInput) {
			Fail(c, 400, err.Error())
			return
		}
		if isWordbookRequestError(err) {
			Fail(c, 400, err.Error())
			return
		}
		log.Printf("[UpdateCycle] error: %v", err)
		Fail(c, 500, "internal server error")
		return
	}

	OK(c, result)
}

// ClearAllCycles handles DELETE /api/cycles/all?userId=xxx&wordbook=KET
func ClearAllCycles(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		Fail(c, 400, "userId is required")
		return
	}
	if !requireAdminUser(c, userID) {
		return
	}
	wordbook := c.Query("wordbook")
	if wordbook == "" {
		Fail(c, 400, "wordbook is required")
		return
	}

	if err := service.ClearAllCycles(userID, wordbook); err != nil {
		if isWordbookRequestError(err) {
			Fail(c, 400, err.Error())
			return
		}
		log.Printf("[ClearAllCycles] error: %v", err)
		Fail(c, 500, "internal server error")
		return
	}

	OK(c, gin.H{"ok": true})
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
	isAdmin, err := accountService.IsAdmin(account.ID)
	if err != nil {
		log.Printf("[Login] isAdmin error: %v", err)
		Fail(c, 500, "internal server error")
		return
	}

	OK(c, gin.H{
		"userId":      account.ID,
		"accountName": account.Name,
		"isAdmin":     isAdmin,
	})
}

// CreateAccountRequest is the request body for POST /api/account/create.
type CreateAccountRequest struct {
	AccountName string `json:"accountName" binding:"required"`
}

// CreateAccount handles POST /api/account/create?userId=xxx.
func CreateAccount(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		Fail(c, 400, "userId is required")
		return
	}
	if !requireAdminUser(c, userID) {
		return
	}

	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "invalid request body")
		return
	}

	accountService := service.AccountService{}
	account, err := accountService.Create(req.AccountName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAccountNameRequired):
			Fail(c, 400, "accountName is required")
		case errors.Is(err, service.ErrAccountAlreadyExists):
			Fail(c, 409, "account already exists")
		default:
			log.Printf("[CreateAccount] error: %v", err)
			Fail(c, 500, "internal server error")
		}
		return
	}

	OK(c, gin.H{
		"userId":      account.ID,
		"accountName": account.Name,
	})
}

func requireAdminUser(c *gin.Context, userID string) bool {
	parsedUserID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil || parsedUserID == 0 {
		Fail(c, 400, "invalid userId")
		return false
	}

	accountService := service.AccountService{}
	isAdmin, err := accountService.IsAdmin(uint(parsedUserID))
	if err != nil {
		log.Printf("[requireAdminUser] error: %v", err)
		Fail(c, 500, "internal server error")
		return false
	}
	if !isAdmin {
		Fail(c, 403, "admin only")
		return false
	}
	return true
}

func isWordbookRequestError(err error) bool {
	return errors.Is(err, wordbook.ErrWordbookNotFound) || errors.Is(err, wordbook.ErrWordbookDisabled)
}
