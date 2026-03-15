package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	adminDomain "github.com/infosec554/clean-archtectura/domain/admin"
)

type JWTManager struct {
	SecretKey []byte
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{SecretKey: []byte(secret)}
}

// GenerateAdmin — admin uchun access va refresh token yaratadi
func (j *JWTManager) GenerateAdmin(admin adminDomain.Admin) (string, string, error) {
	// Access Token (24 soat)
	accessClaims := jwt.MapClaims{
		"admin_id": admin.ID.String(),
		"email":    admin.Email,
		"role":     string(admin.Role),
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(j.SecretKey)
	if err != nil {
		return "", "", err
	}

	// Refresh Token (7 kun)
	refreshClaims := jwt.MapClaims{
		"admin_id": admin.ID.String(),
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(j.SecretKey)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Verify — tokenni tekshiradi va claimslarni qaytaradi
func (j *JWTManager) Verify(tokenStr string) (bool, jwt.MapClaims, error) {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return j.SecretKey, nil
	})
	if err != nil {
		return false, nil, err
	}

	if claims, ok := t.Claims.(jwt.MapClaims); ok && t.Valid {
		return true, claims, nil
	}
	return false, nil, fmt.Errorf("invalid token")
}
