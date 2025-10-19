package handlers

import (
	"log"
	"net/http"

	"github.com/TrueSmartcomm/backend/internal/auth"
	"github.com/TrueSmartcomm/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// AuthHandler структура для хендлеров аутентификации
type AuthHandler struct {
    service *auth.AuthService
}

// NewAuthHandler конструктор для AuthHandler
func NewAuthHandler(service *auth.AuthService) *AuthHandler {
    return &AuthHandler{service: service}
}

// RegisterHandler хендлер для регистрации

func (h *AuthHandler) RegisterHandler(c *gin.Context) {
    //декодировать JSON тело запроса в структуру
    var req models.UserRegisterRequest 
    if err := c.ShouldBindJSON(&req); err != nil {
      
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    //Вызвать метод сервиса для регистрации
    err := h.service.Register(c.Request.Context(), &req)
    if err != nil {
        if err.Error() == "user with this login already exists" || err.Error() == "user with this email already exists" {
            c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
            return
        }
        // Другие ошибки (например, ошибка БД)
        log.Printf("Auth Handler Register error: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
        return
    }

    //Успешно зарегистрирован
    c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// LoginHandler хендлер для логина

func (h *AuthHandler) LoginHandler(c *gin.Context) {
    var req models.UserLoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    resp, err := h.service.Login(c.Request.Context(), &req)
    if err != nil {
        // Важно: возвращаем 401 для неверных данных, а не 400 или 500
        log.Printf("Auth Handler Login error: %v", err) // Логируем для отладки
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login or password"})
        return
    }

    // 4. Успешно аутентифицирован, возвращаем токены
    c.JSON(http.StatusOK, gin.H{
        "token":         resp.Token,
        "refresh_token": resp.RefreshToken, // Включаем refresh токен в ответ
    })
}

// RefreshTokenHandler хендлер для обновления токена

func (h *AuthHandler) RefreshTokenHandler(c *gin.Context) {
    var req models.UserRefreshRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    resp, err := h.service.RefreshToken(c.Request.Context(), &req)
    if err != nil {
        log.Printf("Auth Handler Refresh error: %v", err)
        // Ошибки RefreshToken (не найден, просрочен) возвращаются как 401
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    // 5. Успешно обновлен токен
    c.JSON(http.StatusOK, gin.H{
        "token": resp.Token,
    
    })
}

// ProfileHandler (пример защищенного хендлера)

func (h *AuthHandler) ProfileHandler(c *gin.Context) {
    
    userID, exists := c.Get("user_id") 
    if !exists {
        
        log.Println("Auth Handler Profile: user_id not found in context")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    // Приводим к нужному типу (int, как в ValidateToken)
    userIDInt, ok := userID.(int)
    if !ok {
        log.Println("Auth Handler Profile: user_id is not int")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }


    c.JSON(http.StatusOK, gin.H{
        "message": "Welcome to your profile!",
        "user_id": userIDInt,
    })
}