package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"wordtask-server/api"
)

func main() {
	// 生产环境可改为 gin.ReleaseMode
	gin.SetMode(gin.DebugMode)

	r := gin.Default()

	// 跨域配置：允许 Vite 前端开发端口访问
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// 健康检查
	r.GET("/ping", api.Ping)

	// API 路由分组
	apiGroup := r.Group("/api")
	{
		words := apiGroup.Group("/words")
		{
			words.GET("/today", api.GetTodayWords)
			words.POST("/review", api.SubmitReview)
		}
	}

	addr := ":8080"
	log.Printf("WordTask server running at http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
