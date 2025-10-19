package main

import (
	"log"
	"os"

	"github.com/TrueSmartcomm/backend/config"
	"github.com/TrueSmartcomm/backend/internal/auth"
	"github.com/TrueSmartcomm/backend/internal/handler"
	"github.com/TrueSmartcomm/backend/internal/middleware"
	"github.com/TrueSmartcomm/backend/internal/repository"
	"github.com/TrueSmartcomm/backend/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.MustLoad()

	db, err := storage.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}
	defer db.Close()

	// --- Инициализация зависимостей для аутентификации ---
	userRepo := repository.NewUserRepository(db.DB) // Репозиторий для пользователей

	secretKey := []byte(os.Getenv("SECRET_KEY")) // Получаем ключ из переменной окружения
	if len(secretKey) == 0 {
		log.Fatal("SECRET_KEY environment variable is not set")
	}
	authService := auth.NewAuthService(userRepo, secretKey) // Сервис аутентификации
	// --- Конец инициализации аутентификации ---

	// Инициализация репозиториев и хендлеров для задач
	taskRepo := repository.NewTaskRepository(db.DB)
	taskHandler := handlers.NewTaskHandler(taskRepo) // Хендлер задач

	// Инициализация хендлеров аутентификации
	authHandler := handlers.NewAuthHandler(authService) // Хендлер аутентификации

	// --- Настройка маршрутов ---
	r := gin.New()
	corsConfig := cors.DefaultConfig()

	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
	}

	r.Use(gin.Recovery(), gin.Logger(), cors.New(corsConfig))

	// Healthcheck endpoint (публичный)
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	// Проверка подключения к БД (публичный)
	r.GET("/db-check", func(ctx *gin.Context) {
		var now string
		if err := db.DB.QueryRow(ctx, "SELECT NOW()").Scan(&now); err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(200, gin.H{"db_time": now})
	})

	// --- Публичные маршруты (не требуют аутентификации) ---
	// Регистрация
	r.POST("/api/v1/auth/register", authHandler.RegisterHandler)
	// Логин
	r.POST("/api/v1/auth/login", authHandler.LoginHandler)
	// Обновление токена (публичный маршрут, так как использует refresh токен из тела)
	r.POST("/api/v1/auth/refresh", authHandler.RefreshTokenHandler)

	// --- Защищённые маршруты (требуют аутентификацию через JWT) ---
	// Создаём группу маршрутов с middleware
	authorized := r.Group("/api/v1") // Можно использовать и другой префикс, например /api/v1
	// Применяем middleware ко всей группе
	authorized.Use(middleware.AuthRequired(authService)) // Передаём authService для проверки токена

	{

		authorized.POST("/api/v1/tasks", taskHandler.CreateTask)
		authorized.GET("/api/v1/tasks", taskHandler.GetTask)
		authorized.PUT("/api/v1/tasks", taskHandler.UpdateTask)
		authorized.DELETE("/api/v1/tasks", taskHandler.DeleteTask)
		authorized.POST("/api/v1/tasks/move", taskHandler.MoveTask)
		authorized.POST("/api/v1/tasks/dependency", taskHandler.AddTaskDependency)
		authorized.DELETE("/api/v1/tasks/dependency", taskHandler.RemoveTaskDependency)
		authorized.GET("/api/v1/tasks/with-dependencies", taskHandler.GetTaskWithDependencies)

		authorized.GET("/api/v1/profile", func(c *gin.Context) {
			userID, exists := middleware.GetUserIDFromContext(c)
			if !exists {

				log.Println("Main: user_id not found in context after AuthRequired middleware")
				c.JSON(500, gin.H{"error": "Internal server error"})
				return
			}
			c.JSON(200, gin.H{
				"message": "Welcome to your profile!",
				"user_id": userID,
			})
		})
	}

	// HTTP сервер
	log.Printf("starting http server on :%s...", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
