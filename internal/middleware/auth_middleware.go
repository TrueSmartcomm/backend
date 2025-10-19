package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/TrueSmartcomm/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

// ContextKey тип для ключа user_id в контексте
type ContextKey string

const UserIDKey ContextKey = "user_id" // Ключ для хранения user_id в контексте запроса

// AuthRequired возвращает Gin middleware, которое проверяет JWT токен
func AuthRequired(authService *auth.AuthService) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		//Получить заголовок Authorization
		authHeader := c.GetHeader("Authorization")

		//Проверить, начинается ли он с "Bearer "
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing or invalid"})
			c.Abort() // Прерываем цепочку выполнения
			return
		}

		//Извлечь токен (убираем "Bearer ")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		//Проверить токен с помощью AuthService
		userID, err := authService.ValidateToken(tokenString)
		if err != nil {
			log.Printf("Auth Middleware: Invalid token - %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		//Если токен валиден, положить userID в контекст
		c.Set(string(UserIDKey), userID)
		//Продолжить выполнение цепочки (перейти к следующему middleware или основному хендлеру)
		c.Next()
	})
}

func GetUserIDFromContext(c *gin.Context) (int, bool) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return 0, false
	}

	userIDInt, ok := userID.(int)
	if !ok {
		return 0, false
	}
	return userIDInt, true
}
