package api

import (
	"log"

	"github.com/gin-gonic/gin"

	"wordtask-server/mock"
)

// Ping 健康检查接口
func Ping(c *gin.Context) {
	OK(c, gin.H{"message": "pong"})
}

// GetTodayWords 获取今日待复习的单词列表
// GET /api/words/today
func GetTodayWords(c *gin.Context) {
	OK(c, mock.TodayWords)
}

// ReviewRequest 前端提交的复习评价请求体
type ReviewRequest struct {
	WordID int    `json:"word_id" binding:"required"`
	Grade  string `json:"grade" binding:"required,oneof=red yellow green"`
}

// SubmitReview 提交对某个单词的复习评价
// POST /api/words/review
// Body: {"word_id": 1, "grade": "green"}
func SubmitReview(c *gin.Context) {
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "invalid params: "+err.Error())
		return
	}

	// Mock 阶段：仅打印，后续 M4 将接入 FSRS 算法并持久化
	log.Printf("[Review] word_id=%d, grade=%s", req.WordID, req.Grade)

	OK(c, gin.H{
		"word_id": req.WordID,
		"grade":   req.Grade,
		// 下一次复习时间占位，后续由 FSRS 算法生成
		"next_review_at": nil,
	})
}
