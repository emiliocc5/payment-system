package domain

import "time"

type Balance struct {
	UserID    string
	Available int64
	Reserved  int64
	UpdatedAt time.Time
}
