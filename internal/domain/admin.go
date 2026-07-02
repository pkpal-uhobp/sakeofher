package domain

import "time"

type Admin struct {
	ID         int64
	TelegramID int64
	Username   *string
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type AdminAction struct {
	ID           int64
	AdminID      int64
	TargetUserID *int64
	Action       string
	Details      []byte
	CreatedAt    time.Time
}
