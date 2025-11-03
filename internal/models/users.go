package models

import (
	"time"
)
// UserLoginRequest структура для запроса логина
type UserLoginRequest struct {
    Login    string `json:"login" binding:"required"`    
    Password string `json:"password" binding:"required"` 
}

// UserRegisterRequest структура для запроса регистрации
type UserRegisterRequest struct {
    Login    string `json:"login" binding:"required,min=3,max=50"` // Валидация на стороне Go
    Email    string `json:"email" binding:"required,email"`        // Валидация email
    Password string `json:"password" binding:"required,min=8"`     // Валидация пароля
}

// UserAuthResponse структура для успешного ответа аутентификации/регистрации
type UserAuthResponse struct {
    Token         string `json:"token"`          // JWT токен
    RefreshToken  string `json:"refresh_token"`  // Refresh токен 
}

// UserRefreshRequest структура для запроса обновления токена
type UserRefreshRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

type User struct {
    ID           int       `db:"id" json:"id"` 
    Login        string    `db:"login" json:"login"`
    Email        string    `db:"email" json:"email"`
    PasswordHash string    `db:"password_hash" json:"-"` 
    CreatedAt    time.Time `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}