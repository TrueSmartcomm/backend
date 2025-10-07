// ../../internal/repository/user_repo.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/TrueSmartcomm/backend/internal/models"
	"github.com/jackc/pgx/v5"         // Для работы с PostgreSQL
	"github.com/jackc/pgx/v5/pgconn"  // Для проверки кодов ошибок
	"github.com/jackc/pgx/v5/pgxpool" // Пул соединений
)

// UserRepository структура для работы с пользователями в БД
type UserRepository struct {
	DB *pgxpool.Pool
}

// NewUserRepository конструктор для UserRepository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUser создает нового пользователя
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (login, email, password_hash, created_at, updated_at) 
              VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id, created_at, updated_at`

	// Используем QueryRow, так как RETURNING возвращает одну строку
	err := r.DB.QueryRow(ctx, query, user.Login, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// Проверяем на нарушение уникального ограничения (дубликрующий логин или email)
		// pgx возвращает ошибку с кодом через pgconn.PgError
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" { // unique_violation
				// Проверим, какое именно ограничение нарушено, по имени
				if pgErr.ConstraintName == "users_login_key" || pgErr.ConstraintName == "users_email_key" {
					// Возвращаем общую ошибку, чтобы не раскрывать, что именно дублируется
					return errors.New("user with this login or email already exists")
				}
			}
		}
		return err
	}
	return nil
}

// GetUserByLogin находит пользователя по логину
func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	query := `SELECT id, login, email, password_hash, created_at, updated_at FROM users WHERE login = $1`

	err := r.DB.QueryRow(ctx, query, login).Scan(
		&user.ID, &user.Login, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail находит пользователя по email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, login, email, password_hash, created_at, updated_at FROM users WHERE email = $1`

	err := r.DB.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Login, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// SaveRefreshToken сохраняет refresh-токен в базу данных
func (r *UserRepository) SaveRefreshToken(ctx context.Context, userID int, token string, expiresIn time.Duration) error {
	// expiresIn передаётся как Duration (например, 7*24*time.Hour)
	expiresAt := time.Now().Add(expiresIn)
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`
	_, err := r.DB.Exec(ctx, query, userID, token, expiresAt)
	return err
}

// GetRefreshToken находит refresh-токен в базе данных и возвращает userID и время истечения
func (r *UserRepository) GetRefreshToken(ctx context.Context, token string) (int, time.Time, error) {
	var userID int
	var expiresAt time.Time
	query := `SELECT user_id, expires_at FROM refresh_tokens WHERE token = $1`
	err := r.DB.QueryRow(ctx, query, token).Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, time.Time{}, errors.New("refresh token not found")
		}
		return 0, time.Time{}, err
	}
	return userID, expiresAt, nil
}

// DeleteRefreshToken удаляет refresh-токен из базы данных
func (r *UserRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.DB.Exec(ctx, query, token)
	return err
}
