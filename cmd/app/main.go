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
	r.Use(gin.Recovery(), gin.Logger())

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
		// Маршруты для задач (старые, но теперь защищённые)
		// Обрати внимание: изменились пути с "/api/v1/tasks" на "/tasks"
		// Если хочешь сохранить "/api/v1", измени здесь и в хендлерах
		authorized.POST("/tasks", taskHandler.CreateTask)
		authorized.GET("/tasks", taskHandler.GetTask)
		authorized.PUT("/tasks", taskHandler.UpdateTask)
		authorized.DELETE("/tasks", taskHandler.DeleteTask)
		// Если у тебя есть другие методы в taskHandler, добавь их сюда
		// authorized.POST("/tasks/move", taskHandler.MoveTask) // Если этот метод есть и он должен быть защищён
		// authorized.POST("/tasks/dependency", taskHandler.AddTaskDependency)
		// authorized.DELETE("/tasks/dependency", taskHandler.RemoveTaskDependency)
		// authorized.GET("/tasks/with-dependencies", taskHandler.GetTaskWithDependencies)

		// Пример другого защищенного маршрута (профиль)
		authorized.GET("/profile", func(c *gin.Context) {
			userID, exists := middleware.GetUserIDFromContext(c) // Используем вспомогательную функцию из middleware
			if !exists {
				// Это маловероятно, если middleware работает правильно
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
	// --- Конец настройки маршрутов ---

	// HTTP сервер
	log.Printf("starting http server on :%s...", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
