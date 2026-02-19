package user

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/infosec554/clean-archtectura/config"
	"github.com/infosec554/clean-archtectura/pkg/cache"
	"github.com/infosec554/clean-archtectura/pkg/email"
	"github.com/rs/zerolog"

	domain "github.com/infosec554/clean-archtectura/domain/users"
	"github.com/infosec554/clean-archtectura/pkg/token"
)

const verifyCodeTTL = 5 * time.Minute

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	Create(ctx context.Context, req *domain.CreateUser) (string, error)
	Update(ctx context.Context, req *domain.UpdateUser, passwordHash string) (string, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SetEmailVerified(ctx context.Context, email string) error
}

type UserService struct {
	repo       UserRepository
	cache      cache.ICache
	emailSender *email.Sender
	logger     zerolog.Logger
	jwtManager *token.JWTManager
}

func NewUserService(repo UserRepository, cfg config.Config, c cache.ICache, logger zerolog.Logger, jwtManager *token.JWTManager) *UserService {
	return &UserService{
		repo:        repo,
		cache:       c,
		emailSender: email.NewSender(cfg),
		logger:      logger.With().Str("service", "user").Logger(),
		jwtManager:  jwtManager,
	}
}

// Register creates a new user and sends email verification code
func (s *UserService) Register(ctx context.Context, req *domain.CreateUser) (string, error) {
	id, err := s.repo.Create(ctx, req)
	if err != nil {
		return "", err
	}

	// Verification code yuborish
	if req.Email != "" {
		if sendErr := s.sendCode(req.Email); sendErr != nil {
			s.logger.Warn().Err(sendErr).Str("email", req.Email).Msg("Failed to send verification email")
		}
	}

	return id, nil
}

// SendVerificationCode — resend uchun
func (s *UserService) SendVerificationCode(ctx context.Context, req *domain.ResendCodeRequest) error {
	_, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return errors.New("user not found")
	}
	return s.sendCode(req.Email)
}

// VerifyEmail — codeni tekshirib email_verified=true qiladi
func (s *UserService) VerifyEmail(ctx context.Context, req *domain.VerifyEmailRequest) error {
	key := verifyKey(req.Email)
	stored, err := s.cache.Get(key)
	if err != nil {
		return errors.New("code expired or not found")
	}
	if stored != req.Code {
		return errors.New("invalid verification code")
	}

	if err := s.repo.SetEmailVerified(ctx, req.Email); err != nil {
		return err
	}

	// Codeni o'chirish (TTL=0 qilib)
	_ = s.cache.Set(key, "", time.Millisecond)
	return nil
}

// Login authenticates a user and returns tokens
func (s *UserService) Login(ctx context.Context, req *domain.LoginRequest) (domain.LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return domain.LoginResponse{}, errors.New("invalid credentials")
	}

	if !user.EmailVerified {
		return domain.LoginResponse{}, errors.New("email not verified")
	}

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

// UpdatePassword checks old password and sets new one
func (s *UserService) UpdatePassword(ctx context.Context, userID uuid.UUID, req *domain.UpdatePasswordRequest) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Password == nil || *user.Password != req.OldPassword {
		return errors.New("invalid old password")
	}

	_, err = s.repo.Update(ctx, &domain.UpdateUser{ID: userID, Password: req.NewPassword}, req.NewPassword)
	return err
}

// Delete removes a user
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid user id")
	}
	return s.repo.Delete(ctx, id)
}

// --- helpers ---

func (s *UserService) sendCode(toEmail string) error {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	if err := s.cache.Set(verifyKey(toEmail), code, verifyCodeTTL); err != nil {
		return err
	}
	return s.emailSender.SendVerificationCode(toEmail, code)
}

func verifyKey(email string) string {
	return "verify:" + email
}

func convertToUserResponse(user domain.User) domain.UserResponse {
	resp := domain.UserResponse{
		ID:            user.ID,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
	if user.Email != nil {
		resp.Email = *user.Email
	}
	return resp
}
