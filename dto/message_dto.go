package dto

import "time"

type MessageDto struct {
	ID        string
	Message   string
	From      string
	Timestamp time.Time
}
