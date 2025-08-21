package main

import (
	"log"

	"github.com/TrueSmartcomm/backend/config"
	"github.com/TrueSmartcomm/backend/internal/handler"
	"github.com/TrueSmartcomm/backend/internal/repository"
	"github.com/TrueSmartcomm/backend/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	// загрузкамана конфиг
	cfg := config.MustLoad()

	db, err := storage.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}

	defer db.Close()

	//роутер
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	// Healthcheck endpoint
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	// Пример запроса в БД, обсудить с Егором бд
	r.GET("/db-check", func(ctx *gin.Context) {
		var now string
		if err := db.DB.QueryRow(ctx, "SELECT NOW()").Scan(&now); err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(200, gin.H{"db_time": now})
	})

	taskRepo := repository.NewTaskRepository(db.DB)
	taskHandler := handlers.NewTaskHandler(taskRepo)

	r.POST("/tasks", taskHandler.CreateTask)
	r.GET("/tasks", taskHandler.GetTask)
	r.PUT("/tasks", taskHandler.UpdateTask)
	r.DELETE("/tasks", taskHandler.DeleteTask)

	//HTTP сервак
	log.Printf("starting http server on :%s...", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
