package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	domain "github.com/infosec554/clean-archtectura/domain/users"
)

type UserRepository struct {
	DB     *sql.DB
	logger zerolog.Logger
}

func NewUserRepository(db *sql.DB, logger zerolog.Logger) *UserRepository {
	return &UserRepository{
		DB:     db,
		logger: logger.With().Str("repository", "user").Logger(),
	}
}

// Create ...
func (r *UserRepository) Create(ctx context.Context, req *domain.CreateUser) (string, error) {
	var id uuid.UUID
	query := `
		INSERT INTO users (first_name, last_name, email, password)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.DB.QueryRowContext(ctx, query,
		req.FirstName,
		req.LastName,
		req.Email,
		req.Password,
	).Scan(&id)

	if err != nil {
		r.logger.Error().Err(err).Msg("Error creating user")
		return "", err
	}

	return id.String(), nil
}

// GetByID retrieves a single user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var (
		user      domain.User
		email     sql.NullString
		password  sql.NullString
		createdAt sql.NullTime
		updatedAt sql.NullTime
	)

	query := `
		SELECT id, first_name, last_name, email, password, email_verified, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	if err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&email,
		&password,
		&user.EmailVerified,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn().Str("user_id", id.String()).Msg("User not found")
			return domain.User{}, errors.New("user not found")
		}
		r.logger.Error().Err(err).Str("user_id", id.String()).Msg("Error scanning user by ID")
		return domain.User{}, err
	}

	if email.Valid {
		user.Email = &email.String
	}
	if password.Valid {
		user.Password = &password.String
	}
	if createdAt.Valid {
		user.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}

	return user, nil
}

// SetEmailVerified marks the user's email as verified
func (r *UserRepository) SetEmailVerified(ctx context.Context, email string) error {
	query := `UPDATE users SET email_verified = TRUE, updated_at = NOW() WHERE email = $1`

	result, err := r.DB.ExecContext(ctx, query, email)
	if err != nil {
		r.logger.Error().Err(err).Str("email", email).Msg("Error setting email verified")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	r.logger.Info().Str("email", email).Msg("Email verified successfully")
	return nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, req *domain.UpdateUser, passwordHash string) (string, error) {
	query := `
		UPDATE users
		SET
			first_name = COALESCE(NULLIF($1, ''), first_name),
			last_name = COALESCE(NULLIF($2, ''), last_name),
			email = COALESCE(NULLIF($3, ''), email),
			password = COALESCE(NULLIF($4, ''), password),
			updated_at = NOW()
		WHERE id = $5
	`

	result, err := r.DB.ExecContext(ctx, query,
		req.FirstName,
		req.LastName,
		req.Email,
		passwordHash,
		req.ID,
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", req.ID.String()).Msg("Error updating user")
		return "", err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn().Str("user_id", req.ID.String()).Msg("User not found for update")
		return "", errors.New("user not found")
	}

	r.logger.Info().Str("user_id", req.ID.String()).Msg("User updated successfully")
	return req.ID.String(), nil
}

// Delete removes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", id.String()).Msg("Error deleting user")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn().Str("user_id", id.String()).Msg("User not found for deletion")
		return errors.New("user not found")
	}

	r.logger.Info().Str("user_id", id.String()).Msg("User deleted successfully")
	return nil
}

// GetByEmail retrieves a single user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var (
		user      domain.User
		dbEmail   sql.NullString
		password  sql.NullString
		createdAt sql.NullTime
		updatedAt sql.NullTime
	)

	query := `
		SELECT id, first_name, last_name, email, password, email_verified, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	if err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&dbEmail,
		&password,
		&user.EmailVerified,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn().Str("email", email).Msg("User not found")
			return domain.User{}, errors.New("user not found")
		}
		r.logger.Error().Err(err).Str("email", email).Msg("Error scanning user by email")
		return domain.User{}, err
	}

	if dbEmail.Valid {
		user.Email = &dbEmail.String
	}
	if password.Valid {
		user.Password = &password.String
	}
	if createdAt.Valid {
		user.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}

	return user, nil
}
