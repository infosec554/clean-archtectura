package user

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/infosec554/clean-archtectura/config"
	"github.com/infosec554/clean-archtectura/pkg/cache"

	domain "github.com/infosec554/clean-archtectura/domain/users"
	"github.com/infosec554/clean-archtectura/pkg/token"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	GetUserFirstActiveCompany(ctx context.Context, userID uuid.UUID) (*uuid.UUID, string, error)
	Update(ctx context.Context, req *domain.UpdateUser, passwordHash string) (string, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
}

type UserService struct {
	repo       UserRepository
	logger     zerolog.Logger
	jwtManager *token.JWTManager
}

func NewUserService(repo UserRepository, cfg config.Config, c cache.ICache, logger zerolog.Logger, jwtManager *token.JWTManager) *UserService {
	return &UserService{
		repo:       repo,
		logger:     logger.With().Str("service", "user").Logger(),
		jwtManager: jwtManager,
	}
}

// GetByID retrieves a single user by ID
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (domain.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.UserResponse{}, err
	}

	return convertToUserResponse(user), nil
}

// Update updates an existing user
func (s *UserService) Update(ctx context.Context, req *domain.UpdateUser) (string, error) {
	if req == nil {
		return "", errors.New("empty update request")
	}
	if req.ID == uuid.Nil {
		return "", errors.New("invalid user id")
	}

	return s.repo.Update(ctx, req, "")
}

// Delete removes a user
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid user id")
	}

	return s.repo.Delete(ctx, id)
}

// Helper function to convert domain.User to domain.UserResponse
func convertToUserResponse(user domain.User) domain.UserResponse {
	resp := domain.UserResponse{
		ID:        user.ID,
		Passport:  "",
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if user.Passport != nil {
		resp.Passport = *user.Passport
	}
	if user.Email != nil {
		resp.Email = *user.Email
	}

	resp.FullName = strings.TrimSpace(resp.LastName + " " + resp.FirstName)

	return resp
}
