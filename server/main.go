package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"wordtask-server/api"
	"wordtask-server/internal/db"
	"wordtask-server/internal/wordbook"
)

func main() {
	gin.SetMode(gin.DebugMode)

	// Initialize database (creates file + auto-migrates tables).
	dsn := getEnv("DB_PATH", "./data/wordtask.db")
	if err := db.Init(dsn); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	// Seed wordbook data on first run.
	wordbookConfig := getEnv("WORDBOOK_CONFIG", "./config/wordbooks.json")
	if err := wordbook.LoadRegistry(wordbookConfig); err != nil {
		log.Fatalf("failed to load wordbook config: %v", err)
	}
	wordbookDir := getEnv("WORDBOOK_DIR", "../data/wordbooks")
	if err := db.SeedWords(wordbookDir); err != nil {
		log.Fatalf("failed to seed words: %v", err)
	}

	// Seed accounts data.
	adminAccountName := getEnv("ADMIN_ACCOUNT_NAME", "admin")
	if err := db.SeedAccounts(adminAccountName); err != nil {
		log.Fatalf("failed to seed accounts: %v", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://45.62.109.123:3333"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/ping", api.Ping)

	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/wordbooks", api.ListWordbooks)

		words := apiGroup.Group("/words")
		{
			words.GET("/today", api.GetTodayWords)
			words.POST("/review", api.SubmitReview)
			words.GET("/review/due", api.GetReviewDueWords)
			words.POST("/review/revision", api.SubmitRevision)
		}
		progress := apiGroup.Group("/progress")
		{
			progress.POST("/advance", api.AdvanceProgress)
		}

		account := apiGroup.Group("/account")
		{
			account.POST("/login", api.Login)
		}
		cycles := apiGroup.Group("/cycles")
		{
			cycles.GET("", api.ListCycles)
			cycles.POST("/setup", api.SetupCycle)
			cycles.GET("/current", api.GetCurrentCycle)
			cycles.PUT("/current", api.UpdateCycle)
			cycles.GET("/:id", api.GetCycleDetail)
			cycles.DELETE("/all", api.ClearAllCycles)
		}
	}

	addr := ":3335"
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
