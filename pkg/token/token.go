package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	userDomain "github.com/infosec554/clean-archtectura/domain/users"
)

type JWTManager struct {
	SecretKey []byte
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{SecretKey: []byte(secret)}
}

// Generate generates access and refresh tokens for users
func (j *JWTManager) Generate(user userDomain.User) (string, string, error) {
	// Access Token
	accessExp := time.Now().Add(24 * time.Hour)
	accessClaims := jwt.MapClaims{
		"user_id":    user.ID.String(),
		"email":      derefString(user.Email),
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"exp":        accessExp.Unix(),
		"iat":        time.Now().Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(j.SecretKey)
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshExp := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     refreshExp.Unix(),
		"iat":     time.Now().Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(j.SecretKey)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (j *JWTManager) Verify(tokenStr string) (bool, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return j.SecretKey, nil
	})
	if err != nil {
		return false, nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true, claims, nil
	}
	return false, nil, fmt.Errorf("invalid token")
}

// pointerli stringni xavfsiz ochish
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
