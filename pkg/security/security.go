package security

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", bcrypt.ErrHashTooShort
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPassword(hashedPassword, plainPassword string) bool {
	if hashedPassword == "" || plainPassword == "" {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword)) == nil
}
