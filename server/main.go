package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"wordtask-server/api"
	"wordtask-server/internal/db"
)

func main() {
	gin.SetMode(gin.DebugMode)

	// Initialize database (creates file + auto-migrates tables).
	dsn := getEnv("DB_PATH", "./data/wordtask.db")
	if err := db.Init(dsn); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	// Seed wordbook data on first run.
	wordbookDir := getEnv("WORDBOOK_DIR", "../data/wordbooks")
	if err := db.SeedWords(wordbookDir); err != nil {
		log.Fatalf("failed to seed words: %v", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/ping", api.Ping)

	apiGroup := r.Group("/api")
	{
		words := apiGroup.Group("/words")
		{
			words.GET("/today", api.GetTodayWords)
			words.POST("/review", api.SubmitReview)
		}
		progress := apiGroup.Group("/progress")
		{
			progress.POST("/advance", api.AdvanceProgress)
		}
	}

	addr := ":8080"
	log.Printf("WordTask server running at http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
