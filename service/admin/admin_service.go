package admin

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	domain "github.com/infosec554/clean-archtectura/domain/admin"
	"github.com/infosec554/clean-archtectura/config"
	"github.com/infosec554/clean-archtectura/pkg/cache"
	"github.com/infosec554/clean-archtectura/pkg/email"
	"github.com/infosec554/clean-archtectura/pkg/token"
)

const (
	loginCodeTTL  = 5 * time.Minute
	inviteTokenTTL = 24 * time.Hour
)

type AdminRepository interface {
	GetByEmail(ctx context.Context, email string) (domain.Admin, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Admin, error)
	Create(ctx context.Context, email, passwordHash string, role domain.AdminRole, invitedBy uuid.UUID) (uuid.UUID, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, page int) (domain.AdminList, error)
	SaveInvite(ctx context.Context, email string, token string, invitedBy uuid.UUID, expiresAt time.Time) error
	GetInvite(ctx context.Context, token string) (string, uuid.UUID, error)
	MarkInviteUsed(ctx context.Context, token string) error
}

type AdminService struct {
	repo        AdminRepository
	cache       cache.ICache
	emailSender *email.Sender
	jwtManager  *token.JWTManager
	logger      zerolog.Logger
}

func NewAdminService(repo AdminRepository, cfg config.Config, c cache.ICache, logger zerolog.Logger, jwtManager *token.JWTManager) *AdminService {
	return &AdminService{
		repo:        repo,
		cache:       c,
		emailSender: email.NewSender(cfg),
		jwtManager:  jwtManager,
		logger:      logger.With().Str("service", "admin").Logger(),
	}
}

// Login — email+parol tekshirib, 2FA kod yuboradi
func (s *AdminService) Login(ctx context.Context, req *domain.LoginRequest) error {
	admin, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return errors.New("invalid credentials")
	}

	if !admin.IsActive {
		return errors.New("account is not active")
	}

	if admin.Password != req.Password {
		return errors.New("invalid credentials")
	}

	// 2FA kod yuborish
	return s.sendLoginCode(req.Email)
}

// VerifyLogin — 2FA kodni tekshirib JWT qaytaradi
func (s *AdminService) VerifyLogin(ctx context.Context, req *domain.VerifyLoginRequest) (domain.LoginResponse, error) {
	key := loginKey(req.Email)
	stored, err := s.cache.Get(key)
	if err != nil {
		return domain.LoginResponse{}, errors.New("code expired or not found")
	}
	if stored != req.Code {
		return domain.LoginResponse{}, errors.New("invalid verification code")
	}

	// Kodni o'chirish
	_ = s.cache.Set(key, "", time.Millisecond)

	admin, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return domain.LoginResponse{}, err
	}

	accessToken, refreshToken, err := s.jwtManager.GenerateAdmin(admin)
	if err != nil {
		return domain.LoginResponse{}, err
	}

	return domain.LoginResponse{
		Admin:        toResponse(admin),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// InviteAdmin — superadmin yangi admin taklif qiladi
func (s *AdminService) InviteAdmin(ctx context.Context, req *domain.InviteAdminRequest, invitedBy uuid.UUID) error {
	// Email allaqachon mavjudmi tekshirish
	if _, err := s.repo.GetByEmail(ctx, req.Email); err == nil {
		return errors.New("admin with this email already exists")
	}

	inviteToken := fmt.Sprintf("%06d", rand.Intn(1000000))
	expiresAt := time.Now().Add(inviteTokenTTL)

	if err := s.repo.SaveInvite(ctx, req.Email, inviteToken, invitedBy, expiresAt); err != nil {
		return err
	}

	// Invite tokenini Redisga ham saqlaymiz (tez tekshirish uchun)
	_ = s.cache.Set("invite:"+inviteToken, req.Email, inviteTokenTTL)

	return s.emailSender.SendInviteCode(req.Email, inviteToken)
}

// AcceptInvite — yangi admin tokenni qabul qilib parol o'rnatadi
func (s *AdminService) AcceptInvite(ctx context.Context, req *domain.AcceptInviteRequest) error {
	emailAddr, invitedBy, err := s.repo.GetInvite(ctx, req.Token)
	if err != nil {
		return err
	}

	if _, err := s.repo.Create(ctx, emailAddr, req.Password, domain.RoleAdmin, invitedBy); err != nil {
		return err
	}

	return s.repo.MarkInviteUsed(ctx, req.Token)
}

// GetByID — admin ma'lumotlari
func (s *AdminService) GetByID(ctx context.Context, id uuid.UUID) (domain.AdminResponse, error) {
	admin, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.AdminResponse{}, err
	}
	return toResponse(admin), nil
}

// List — adminlar ro'yxati
func (s *AdminService) List(ctx context.Context, limit, page int) (domain.AdminList, error) {
	return s.repo.List(ctx, limit, page)
}

// UpdatePassword — parolni yangilash
func (s *AdminService) UpdatePassword(ctx context.Context, id uuid.UUID, req *domain.UpdatePasswordRequest) error {
	admin, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if admin.Password != req.OldPassword {
		return errors.New("invalid old password")
	}

	return s.repo.UpdatePassword(ctx, id, req.NewPassword)
}

// Delete — admin o'chirish (faqat superadmin)
func (s *AdminService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// --- helpers ---

func (s *AdminService) sendLoginCode(toEmail string) error {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	if err := s.cache.Set(loginKey(toEmail), code, loginCodeTTL); err != nil {
		return err
	}
	return s.emailSender.SendVerificationCode(toEmail, code)
}

func loginKey(email string) string {
	return "admin_login:" + email
}

func toResponse(a domain.Admin) domain.AdminResponse {
	return domain.AdminResponse{
		ID:        a.ID,
		Email:     a.Email,
		Role:      a.Role,
		InvitedBy: a.InvitedBy,
		IsActive:  a.IsActive,
		CreatedAt: a.CreatedAt,
	}
}
