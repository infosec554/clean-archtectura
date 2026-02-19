package students

import "github.com/google/uuid"

type Student struct {
	ID      uuid.UUID `json:"id"`
	JSHSHIR *string   `json:"jshshir"`
	Email   string    `json:"email"`
}
