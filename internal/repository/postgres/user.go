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
		SELECT id, first_name, last_name, email, password, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	if err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&email,
		&password,
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
