package bot

import "github.com/google/uuid"

type BotAuthInfo struct {
	StudentID uuid.UUID `json:"student_id"`
	PINFL     string    `json:"pinfl"`
}
