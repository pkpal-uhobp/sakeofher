package domain

import "time"

type Broadcast struct {
	ID          int64
	AdminID     int64
	Status      BroadcastStatus
	Text        string
	TotalCount  int
	SentCount   int
	FailedCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BroadcastRecipient struct {
	ID          int64
	BroadcastID int64
	UserID      int64
	Status      string
	ErrorText   *string
	SentAt      *time.Time
	CreatedAt   time.Time
}
