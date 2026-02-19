package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/infosec554/clean-archtectura/config"
	"github.com/infosec554/clean-archtectura/pkg/cache"
	"github.com/rs/zerolog"

	domain "github.com/infosec554/clean-archtectura/domain/users"
	"github.com/infosec554/clean-archtectura/pkg/token"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	Create(ctx context.Context, req *domain.CreateUser) (string, error)
	Update(ctx context.Context, req *domain.UpdateUser, passwordHash string) (string, error)
	Delete(ctx context.Context, id uuid.UUID) error
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

// Register creates a new user
func (s *UserService) Register(ctx context.Context, req *domain.CreateUser) (string, error) {
	// Add password hashing here in production
	return s.repo.Create(ctx, req)
}

// Login authenticates a user and returns tokens
func (s *UserService) Login(ctx context.Context, req *domain.LoginRequest) (domain.LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return domain.LoginResponse{}, err
	}

	// Simple password check (in production use bcrypt)
	if user.Password == nil || *user.Password != req.Password {
		return domain.LoginResponse{}, errors.New("invalid credentials")
	}

	accessToken, refreshToken, err := s.jwtManager.Generate(user)
	if err != nil {
		return domain.LoginResponse{}, err
	}

	return domain.LoginResponse{
		User:         convertToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
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

// UpdatePassword ...
func (s *UserService) UpdatePassword(ctx context.Context, userID uuid.UUID, req *domain.UpdatePasswordRequest) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Password == nil || *user.Password != req.OldPassword {
		return errors.New("invalid old password")
	}

	updateReq := &domain.UpdateUser{
		ID:       userID,
		Password: req.NewPassword,
	}

	_, err = s.repo.Update(ctx, updateReq, req.NewPassword)
	return err
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
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if user.Email != nil {
		resp.Email = *user.Email
	}

	return resp
}
