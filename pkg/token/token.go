package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/infosec554/clean-archtectura/domain/bot"
	studentDomain "github.com/infosec554/clean-archtectura/domain/students"
	userDomain "github.com/infosec554/clean-archtectura/domain/users"
)

type JWTManager struct {
	SecretKey []byte
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{SecretKey: []byte(secret)}
}

func (j *JWTManager) GenerateAccessToken(st studentDomain.Student) (string, int64, error) {
	exp := time.Now().AddDate(0, 1, 0) // 1 oy amal qiladi

	claims := jwt.MapClaims{
		"user_id":   st.ID.String(),
		"user_type": "student",
		"pinfl":     derefString(st.JSHSHIR),
		"email":     st.Email,
		"exp":       exp.Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.SecretKey)
	if err != nil {
		return "", 0, err
	}
	return signed, int64(time.Until(exp).Seconds()), nil
}

func (j *JWTManager) GenerateRefreshToken(st studentDomain.Student) (string, error) {
	exp := time.Now().Add(7 * 24 * time.Hour) // 7 kun amal qiladi

	claims := jwt.MapClaims{
		"user_id":   st.ID.String(),
		"user_type": "student",
		"jshshir":   derefString(st.JSHSHIR),
		"email":     st.Email,
		"exp":       exp.Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SecretKey)
}

// GenerateUserAccessToken generates access token for regular users
func (j *JWTManager) GenerateUserAccessToken(user userDomain.User, companyID uuid.UUID) (string, int64, error) {
	exp := time.Now().AddDate(0, 1, 0) // 1 hour validity

	claims := jwt.MapClaims{
		"user_id":    user.ID.String(),
		"user_type":  "user",
		"email":      derefString(user.Email),
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"company_id": companyID.String(),
		"exp":        exp.Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.SecretKey)
	if err != nil {
		return "", 0, err
	}
	return signed, int64(time.Until(exp).Seconds()), nil
}

// GenerateUserRefreshToken generates refresh token for regular users
func (j *JWTManager) GenerateUserRefreshToken(user userDomain.User, companyID uuid.UUID) (string, error) {
	exp := time.Now().Add(7 * 24 * time.Hour) // 7 days validity

	claims := jwt.MapClaims{
		"user_id":    user.ID.String(),
		"user_type":  "user",
		"email":      derefString(user.Email),
		"company_id": companyID.String(),
		"exp":        exp.Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SecretKey)
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

//******************************************************************************

func (j *JWTManager) GenerateBotToken(info bot.BotAuthInfo) (string, error) {

	claims := jwt.MapClaims{
		"user_id": info.StudentID.String(), // ✔ yagona ID
		"pinfl":   info.PINFL,
		"bot":     true,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SecretKey)
}

func (j *JWTManager) ParseBotToken(tokenStr string) (bot.BotAuthInfo, error) {

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return j.SecretKey, nil
	})

	if err != nil || !token.Valid {
		return bot.BotAuthInfo{}, fmt.Errorf("invalid bot token: %w", err)
	}

	claims := token.Claims.(jwt.MapClaims)

	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return bot.BotAuthInfo{}, fmt.Errorf("invalid user_id in token")
	}

	return bot.BotAuthInfo{
		StudentID: userID,
		PINFL:     claims["pinfl"].(string),
	}, nil
}
