package models

import "time"

type Image struct {
	ID           int64      `json:"id"`
	Filename     string     `json:"filename"`
	OriginalName string     `json:"original_name"`
	UploaderID   int64      `json:"uploader_id"`
	CategoryID   *int64     `json:"category_id"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
}
