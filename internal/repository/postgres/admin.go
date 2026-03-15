package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	domain "github.com/infosec554/clean-archtectura/domain/admin"
)

type AdminRepository struct {
	DB     *sql.DB
	logger zerolog.Logger
}

func NewAdminRepository(db *sql.DB, logger zerolog.Logger) *AdminRepository {
	return &AdminRepository{
		DB:     db,
		logger: logger.With().Str("repository", "admin").Logger(),
	}
}

// GetByEmail — email orqali admin topish
func (r *AdminRepository) GetByEmail(ctx context.Context, email string) (domain.Admin, error) {
	var (
		admin     domain.Admin
		invitedBy uuid.NullUUID
		createdAt sql.NullTime
		updatedAt sql.NullTime
	)

	query := `
		SELECT id, email, password, role, invited_by, is_active, created_at, updated_at
		FROM admins
		WHERE email = $1
	`

	if err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&admin.ID,
		&admin.Email,
		&admin.Password,
		&admin.Role,
		&invitedBy,
		&admin.IsActive,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Admin{}, errors.New("admin not found")
		}
		return domain.Admin{}, err
	}

	if invitedBy.Valid {
		admin.InvitedBy = &invitedBy.UUID
	}
	if createdAt.Valid {
		admin.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		admin.UpdatedAt = updatedAt.Time
	}

	return admin, nil
}

// GetByID — ID orqali admin topish
func (r *AdminRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Admin, error) {
	var (
		admin     domain.Admin
		invitedBy uuid.NullUUID
		createdAt sql.NullTime
		updatedAt sql.NullTime
	)

	query := `
		SELECT id, email, password, role, invited_by, is_active, created_at, updated_at
		FROM admins
		WHERE id = $1
	`

	if err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&admin.ID,
		&admin.Email,
		&admin.Password,
		&admin.Role,
		&invitedBy,
		&admin.IsActive,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Admin{}, errors.New("admin not found")
		}
		return domain.Admin{}, err
	}

	if invitedBy.Valid {
		admin.InvitedBy = &invitedBy.UUID
	}
	if createdAt.Valid {
		admin.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		admin.UpdatedAt = updatedAt.Time
	}

	return admin, nil
}

// Create — yangi admin yaratish (invite qabul qilingandan keyin)
func (r *AdminRepository) Create(ctx context.Context, email, passwordHash string, role domain.AdminRole, invitedBy uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID

	query := `
		INSERT INTO admins (email, password, role, invited_by, is_active)
		VALUES ($1, $2, $3, $4, TRUE)
		RETURNING id
	`

	err := r.DB.QueryRowContext(ctx, query, email, passwordHash, role, invitedBy).Scan(&id)
	if err != nil {
		r.logger.Error().Err(err).Str("email", email).Msg("Error creating admin")
		return uuid.Nil, err
	}

	return id, nil
}

// UpdatePassword — parolni yangilash
func (r *AdminRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `UPDATE admins SET password = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.DB.ExecContext(ctx, query, passwordHash, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("admin not found")
	}
	return nil
}

// Delete — admin o'chirish
func (r *AdminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM admins WHERE id = $1`

	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("admin not found")
	}
	return nil
}

// List — barcha adminlar
func (r *AdminRepository) List(ctx context.Context, limit, page int) (domain.AdminList, error) {
	offset := (page - 1) * limit
	var count int

	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(1) FROM admins`).Scan(&count); err != nil {
		return domain.AdminList{}, err
	}

	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, email, role, invited_by, is_active, created_at
		FROM admins
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return domain.AdminList{}, err
	}
	defer rows.Close()

	var list []domain.AdminResponse
	for rows.Next() {
		var a domain.AdminResponse
		var invitedBy uuid.NullUUID
		var createdAt sql.NullTime

		if err := rows.Scan(&a.ID, &a.Email, &a.Role, &invitedBy, &a.IsActive, &createdAt); err != nil {
			return domain.AdminList{}, err
		}
		if invitedBy.Valid {
			a.InvitedBy = &invitedBy.UUID
		}
		if createdAt.Valid {
			a.CreatedAt = createdAt.Time
		}
		list = append(list, a)
	}

	pageCount := (count + limit - 1) / limit
	if pageCount < 1 {
		pageCount = 1
	}

	return domain.AdminList{
		List: list,
		Meta: domain.Meta{Total: count, Page: page, PageSize: limit, PageCount: pageCount},
	}, nil
}

// --- Admin Invites ---

// SaveInvite — invite tokenini saqlash
func (r *AdminRepository) SaveInvite(ctx context.Context, email string, token string, invitedBy uuid.UUID, expiresAt time.Time) error {
	query := `
		INSERT INTO admin_invites (email, token, invited_by, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
	`
	_, err := r.DB.ExecContext(ctx, query, email, token, invitedBy, expiresAt)
	return err
}

// GetInvite — token orqali invite topish
func (r *AdminRepository) GetInvite(ctx context.Context, token string) (string, uuid.UUID, error) {
	var (
		email     string
		invitedBy uuid.UUID
		used      bool
		expiresAt time.Time
	)

	query := `SELECT email, invited_by, used, expires_at FROM admin_invites WHERE token = $1`

	if err := r.DB.QueryRowContext(ctx, query, token).Scan(&email, &invitedBy, &used, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", uuid.Nil, errors.New("invite not found")
		}
		return "", uuid.Nil, err
	}

	if used {
		return "", uuid.Nil, errors.New("invite already used")
	}
	if time.Now().After(expiresAt) {
		return "", uuid.Nil, errors.New("invite expired")
	}

	return email, invitedBy, nil
}

// MarkInviteUsed — inviteni ishlatilgan deb belgilash
func (r *AdminRepository) MarkInviteUsed(ctx context.Context, token string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE admin_invites SET used = TRUE WHERE token = $1`, token)
	return err
}
