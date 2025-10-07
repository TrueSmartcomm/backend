package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/TrueSmartcomm/backend/internal/models"
	"github.com/TrueSmartcomm/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService содержит зависимости для бизнес-логики аутентификации
type AuthService struct {
	repo      *repository.UserRepository // Репозиторий для работы с пользователями в БД
	secretKey []byte                     // Ключ для подписи JWT
}

// NewAuthService конструктор для AuthService
func NewAuthService(repo *repository.UserRepository, secretKey []byte) *AuthService {
	return &AuthService{
		repo:      repo,
		secretKey: secretKey,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, req *models.UserRegisterRequest) error {

	// Проверим, не существует ли пользователь с таким логином или email
	_, err := s.repo.GetUserByLogin(ctx, req.Login)
	if err == nil {
		// Если err == nil, значит пользователь найден
		return errors.New("user with this login already exists")
	}
	// Проверим email
	_, err = s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return errors.New("user with this email already exists")
	}

	//Хеширование пароля
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &models.User{
		Login:        req.Login,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	// 4. Сохранение пользователя в БД через репозиторий
	return s.repo.CreateUser(ctx, user)
}

// Login аутентифицирует пользователя и возвращает токены
func (s *AuthService) Login(ctx context.Context, req *models.UserLoginRequest) (*models.UserAuthResponse, error) {
	//Получение пользователя из БД по логину
	user, err := s.repo.GetUserByLogin(ctx, req.Login)
	if err != nil {

		if err.Error() == "user not found" {

			return nil, errors.New("invalid login or password")
		}
		// Другая ошибка БД
		return nil, err
	}

	//Сравнение введенного пароля с хешем из БД
	err = ComparePassword(user.PasswordHash, req.Password)
	if err != nil {
		// Пароль не совпадает
		return nil, errors.New("invalid login or password")
	}

	//Генерация JWT Access Token
	accessToken, err := s.generateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	//Генерация Refresh Token
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	//Сохранение Refresh Token в БД
	// 7 дней = 7 * 24 * 3600 секунд
	err = s.repo.SaveRefreshToken(ctx, user.ID, refreshToken, 7*24*time.Hour)
	if err != nil {

		return nil, errors.New("failed to save refresh token")
	}

	//Возврат токенов
	return &models.UserAuthResponse{
		Token:        accessToken,
		RefreshToken: refreshToken, // Возвращаем refresh токен при логине
	}, nil
}

// RefreshToken обновляет Access Token, используя Refresh Token
// При ротации старый refresh token удаляется, и сохраняется новый.
func (s *AuthService) RefreshToken(ctx context.Context, req *models.UserRefreshRequest) (*models.UserAuthResponse, error) {
	// 1. Найти refresh token в БД
	userID, expiresAt, err := s.repo.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		// Если токен не найден в БД
		if err.Error() == "refresh token not found" {
			return nil, errors.New("invalid or expired refresh token")
		}
		// Другая ошибка БД
		return nil, err
	}

	// 2. Проверить срок действия
	if time.Now().After(expiresAt) {
		// Удалить просроченный токен (опционально, но рекомендуется)
		s.repo.DeleteRefreshToken(ctx, req.RefreshToken) // Не критично, если ошибка при удалении
		return nil, errors.New("refresh token has expired")
	}

	// 3. (Ротация) Сгенерировать новый Refresh Token и сохранить его.
	// При ротации старый токен удаляется, и новый сохраняется с новым сроком действия.
	// Это повышает безопасность.
	newRefreshToken, err := GenerateRefreshToken()
	if err != nil {
		// Ошибка при генерации нового refresh токена
		return nil, err
	}

	// Удалить старый refresh token из БД
	err = s.repo.DeleteRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		// Логируем ошибку, но не обязательно возвращать её пользователю сразу
		// В зависимости от требований, можно вернуть ошибку или Proceed without deleting old token
		// Здесь возвращаем ошибку, чтобы убедиться, что старый токен удален.
		return nil, errors.New("failed to delete old refresh token")
	}

	// Сохранить новый refresh token в БД
	// 7 дней = 7 * 24 * 3600 секунд
	err = s.repo.SaveRefreshToken(ctx, userID, newRefreshToken, 7*24*time.Hour)
	if err != nil {
		// Логируем ошибку, но не обязательно возвращать её пользователю сразу
		// В зависимости от требований, можно вернуть ошибку или Proceed without saving new token
		// Здесь возвращаем ошибку, чтобы убедиться, что новый токен сохранён.
		return nil, errors.New("failed to save new refresh token")
	}

	// 4. Сгенерировать новый Access Token (JWT) для найденного userID
	newAccessToken, err := s.generateJWT(userID)
	if err != nil {
		// Ошибка при генерации JWT (например, проблема с секретным ключом)
		return nil, err
	}

	// 5. Возврат нового Access Token и НОВОГО Refresh Token
	// В ответе возвращаем новый access токен и новый refresh токен.
	return &models.UserAuthResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken, // Включить, так как используется ротация
	}, nil
}

// generateJWT генерирует JWT токен для пользователя
func (s *AuthService) generateJWT(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), //Токен действителен 24 часа

	})

	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ValidateToken проверяет JWT и возвращает user_id
func (s *AuthService) ValidateToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})

	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64) // JWT числа как float64
	if !ok {
		return 0, errors.New("user_id not found in token claims")
	}

	return int(userIDFloat), nil
}

// HashPassword хеширует пароль
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// ComparePassword сравнивает хэшированный пароль с исходным паролем
func ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateRefreshToken генерирует случайный refresh-токен
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32) // 32 байта = 256 бит
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
