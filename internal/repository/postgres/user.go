package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
		user       domain.User
		middleName sql.NullString
		email      sql.NullString
		phone      sql.NullString
		password   sql.NullString
		createdAt  sql.NullTime
		updatedAt  sql.NullTime
	)

	query := `
		SELECT id, passport, first_name, last_name, email, password, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	if err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Passport,
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

	if middleName.Valid {
		user.MiddleName = &middleName.String
	}
	if email.Valid {
		user.Email = &email.String
	}
	if phone.Valid {
		user.Phone = &phone.String
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

// GetUserFirstActiveCompany retrieves the first active company for a user
func (r *UserRepository) GetUserFirstActiveCompany(ctx context.Context, userID uuid.UUID) (*uuid.UUID, string, error) {
	var (
		companyID uuid.UUID
		userType  string
	)

	query := `
		SELECT company_id, type
		FROM company_users
		WHERE user_id = $1 AND status = 'active'
		ORDER BY company_id ASC
		LIMIT 1
	`

	err := r.DB.QueryRowContext(ctx, query, userID).Scan(&companyID, &userType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Debug().Str("user_id", userID.String()).Msg("User has no active company")
			return nil, "", nil // No active company, not an error
		}
		r.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Error getting user's active company")
		return nil, "", err
	}

	r.logger.Debug().Str("user_id", userID.String()).Str("company_id", companyID.String()).Msg("Found user's active company")
	return &companyID, userType, nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, req *domain.UpdateUser, passwordHash string) (string, error) {
	query := `
		UPDATE users
		SET
			pinfl = COALESCE(NULLIF($1, ''), pinfl),
			passport = COALESCE(NULLIF($2, ''), passport),
			first_name = COALESCE(NULLIF($3, ''), first_name),
			last_name = COALESCE(NULLIF($4, ''), last_name),
			middle_name = COALESCE(NULLIF($5, ''), middle_name),
			email = COALESCE(NULLIF($6, ''), email),
			phone = COALESCE(NULLIF($7, ''), phone),
			password = COALESCE(NULLIF($8, ''), password),
			updated_at = NOW()
		WHERE id = $9
	`

	result, err := r.DB.ExecContext(ctx, query,
		req.PINFL,
		req.Passport,
		req.FirstName,
		req.LastName,
		req.MiddleName,
		req.Email,
		req.Phone,
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

// GetCompanyUsers retrieves all users for a specific company
func (r *UserRepository) GetCompanyUsers(ctx context.Context, companyID uuid.UUID, limit, page int) (domain.CompanyUserList, error) {
	var (
		companyUsers = []domain.CompanyUserResponse{}
		count        = 0
		offset       = (page - 1) * limit
	)

	countQuery := `SELECT COUNT(1) FROM company_users WHERE company_id = $1`

	if err := r.DB.QueryRowContext(ctx, countQuery, companyID).Scan(&count); err != nil {
		r.logger.Error().Err(err).Str("company_id", companyID.String()).Msg("Error counting company users")
		return domain.CompanyUserList{}, err
	}

	query := `
		SELECT
			cu.company_id, cu.user_id, cu.role_id, cu.status,
			u.pinfl, u.passport, u.first_name, u.last_name, u.middle_name, u.email, u.phone, u.created_at, u.updated_at,
			r.id, r.title, r.description, r.created_at, r.updated_at
		FROM company_users cu
		INNER JOIN users u ON cu.user_id = u.id
		INNER JOIN roles r ON cu.role_id = r.id
		WHERE cu.company_id = $1
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.DB.QueryContext(ctx, query, companyID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Str("company_id", companyID.String()).Int("limit", limit).Int("offset", offset).Msg("Error querying company users")
		return domain.CompanyUserList{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cu                                             domain.CompanyUserResponse
			passport, userMiddleName, userEmail, userPhone sql.NullString
			roleDescription                                sql.NullString
			userCreatedAt, userUpdatedAt                   sql.NullTime
			roleCreatedAt, roleUpdatedAt                   sql.NullTime
		)

		if err := rows.Scan(
			&cu.CompanyID,
			&cu.User.ID,
			&cu.Role.ID,
			&cu.Status,
			&cu.User.PINFL,
			&passport,
			&cu.User.FirstName,
			&cu.User.LastName,
			&userMiddleName,
			&userEmail,
			&userPhone,
			&userCreatedAt,
			&userUpdatedAt,
			&cu.Role.ID,
			&cu.Role.Title,
			&roleDescription,
			&roleCreatedAt,
			&roleUpdatedAt,
		); err != nil {
			r.logger.Error().Err(err).Msg("Error scanning company user row")
			return domain.CompanyUserList{}, err
		}

		if passport.Valid {
			cu.User.Passport = passport.String
		}
		if userMiddleName.Valid {
			cu.User.MiddleName = userMiddleName.String
		}
		if userEmail.Valid {
			cu.User.Email = userEmail.String
		}
		if userPhone.Valid {
			cu.User.Phone = userPhone.String
		}
		if userCreatedAt.Valid {
			cu.User.CreatedAt = userCreatedAt.Time
		}
		if userUpdatedAt.Valid {
			cu.User.UpdatedAt = userUpdatedAt.Time
		}
		if roleDescription.Valid {
			cu.Role.Description = roleDescription.String
		}
		if roleCreatedAt.Valid {
			cu.Role.CreatedAt = roleCreatedAt.Time
		}
		if roleUpdatedAt.Valid {
			cu.Role.UpdatedAt = roleUpdatedAt.Time
		}

		cu.User.FullName = fmt.Sprintf("%s %s", cu.User.FirstName, cu.User.LastName)
		if cu.User.MiddleName != "" {
			cu.User.FullName = fmt.Sprintf("%s %s %s", cu.User.FirstName, cu.User.MiddleName, cu.User.LastName)
		}

		companyUsers = append(companyUsers, cu)
	}

	pageCount := (count + limit - 1) / limit
	if pageCount < 1 {
		pageCount = 1
	}

	return domain.CompanyUserList{
		List: companyUsers,
		Meta: domain.Meta{
			Total:     count,
			Page:      page,
			PageSize:  limit,
			PageCount: pageCount,
		},
	}, nil
}

// AssignUserToCompany assigns a user to a company with a role
func (r *UserRepository) AssignUserToCompany(ctx context.Context, req *domain.CreateCompanyUser) error {
	companyID, _ := uuid.Parse(req.CompanyID)
	userID, _ := uuid.Parse(req.UserID)
	roleID, _ := uuid.Parse(req.RoleID)

	query := `
		INSERT INTO company_users (company_id, user_id, role_id, status, type)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (company_id, user_id) DO UPDATE SET role_id = $3, status = $4
	`

	if _, err := r.DB.ExecContext(ctx, query, companyID, userID, roleID, req.Status, req.Type); err != nil {
		r.logger.Error().Err(err).Str("company_id", companyID.String()).Str("user_id", userID.String()).Str("role_id", roleID.String()).Msg("Error assigning user to company")
		return err
	}

	r.logger.Info().Str("company_id", companyID.String()).Str("user_id", userID.String()).Str("role_id", roleID.String()).Str("status", req.Status).Msg("User assigned to company successfully")
	return nil
}

// UpdateCompanyUser updates a user's role or status in a company
func (r *UserRepository) UpdateCompanyUser(ctx context.Context, req *domain.UpdateCompanyUser) error {
	query := `
		UPDATE company_users
		SET
			role_id = COALESCE(NULLIF($1, '00000000-0000-0000-0000-000000000000'::uuid), role_id),
			status = COALESCE(NULLIF($2, ''), status)
		WHERE company_id = $3 AND user_id = $4
	`

	var roleID uuid.UUID
	if req.RoleID != "" {
		roleID, _ = uuid.Parse(req.RoleID)
	}

	result, err := r.DB.ExecContext(ctx, query, roleID, req.Status, req.CompanyID, req.UserID)
	if err != nil {
		r.logger.Error().Err(err).Str("company_id", req.CompanyID.String()).Str("user_id", req.UserID.String()).Msg("Error updating company user")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn().Str("company_id", req.CompanyID.String()).Str("user_id", req.UserID.String()).Msg("Company user not found for update")
		return errors.New("company user not found")
	}

	r.logger.Info().Str("company_id", req.CompanyID.String()).Str("user_id", req.UserID.String()).Msg("Company user updated successfully")
	return nil
}

func (r *UserRepository) RemoveUserFromCompany(ctx context.Context, companyID, userID uuid.UUID) error {
	query := `DELETE FROM company_users WHERE company_id = $1 AND user_id = $2`

	result, err := r.DB.ExecContext(ctx, query, companyID, userID)
	if err != nil {
		r.logger.Error().Err(err).Str("company_id", companyID.String()).Str("user_id", userID.String()).Msg("Error removing user from company")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn().Str("company_id", companyID.String()).Str("user_id", userID.String()).Msg("Company user not found for removal")
		return errors.New("company user not found")
	}

	r.logger.Info().Str("company_id", companyID.String()).Str("user_id", userID.String()).Msg("User removed from company successfully")
	return nil
}

// GetUserCompanyWithRolesAndPermissions retrieves a specific company for a user with their role and permissions
func (r *UserRepository) GetUserCompanyWithRolesAndPermissions(ctx context.Context, userID, companyID uuid.UUID) (domain.MeCompanyResponse, error) {
	query := `
		SELECT
			cu.company_id,
			coalesce(c.name, ue.name) as company_name,
			cu.status,
			r.id as role_id,
			r.title as role_title,
			r.description as role_description,
			p.category,
			p.entity,
			p.code,
			cu.type
		FROM company_users cu
		LEFT JOIN companies c ON cu.company_id = c.id
		LEFT JOIN university_entities ue ON cu.company_id = ue.id
		INNER JOIN roles r ON cu.role_id = r.id
		LEFT JOIN role_permissions rp ON r.id = rp.role_id AND rp.enabled = true
		LEFT JOIN permissions p ON rp.permission_id = p.id
		WHERE cu.user_id = $1 AND cu.company_id = $2
		ORDER BY p.category, p.entity, p.code
	`

	rows, err := r.DB.QueryContext(ctx, query, userID, companyID)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.String()).Str("company_id", companyID.String()).Msg("Error querying user company with roles and permissions")
		return domain.MeCompanyResponse{}, err
	}
	defer rows.Close()

	var company domain.MeCompanyResponse
	permissions := []string{}
	isFirstRow := true

	for rows.Next() {
		var (
			compID          uuid.UUID
			companyName     string
			status          string
			userType        string
			roleID          uuid.UUID
			roleTitle       string
			roleDescription sql.NullString
			category        sql.NullString
			entity          sql.NullString
			code            sql.NullString
		)

		if err := rows.Scan(
			&compID,
			&companyName,
			&status,
			&roleID,
			&roleTitle,
			&roleDescription,
			&category,
			&entity,
			&code,
			&userType,
		); err != nil {
			r.logger.Error().Err(err).Msg("Error scanning company with role and permissions")
			return domain.MeCompanyResponse{}, err
		}

		// Initialize company data on first row
		if isFirstRow {
			company = domain.MeCompanyResponse{
				CompanyID:   compID,
				CompanyName: companyName,
				Status:      status,
				Role: domain.RoleInfo{
					ID:    roleID,
					Title: roleTitle,
				},
				Permissions: []string{},
				UserType:    userType,
			}
			if roleDescription.Valid {
				company.Role.Description = roleDescription.String
			}
			isFirstRow = false
		}

		// Add permission if all parts are present
		if category.Valid && entity.Valid && code.Valid {
			permissionStr := fmt.Sprintf("%s.%s.%s", category.String, entity.String, code.String)
			permissions = append(permissions, permissionStr)
		}
	}

	// Check if company was found
	if isFirstRow {
		r.logger.Warn().Str("user_id", userID.String()).Str("company_id", companyID.String()).Msg("User company not found")
		return domain.MeCompanyResponse{}, errors.New("user company not found")
	}

	company.Permissions = permissions
	r.logger.Debug().Str("user_id", userID.String()).Str("company_id", companyID.String()).Int("permissions_count", len(permissions)).Msg("Retrieved user company with roles and permissions")
	return company, nil
}

// UpdateLastLogin updates the last_login_at timestamp for a company user
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE company_users SET last_login_at = NOW() WHERE user_id = $1`
	_, err := r.DB.ExecContext(ctx, query, userID)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to update last login")
		return err
	}
	return nil
}

// Helper function to convert string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
