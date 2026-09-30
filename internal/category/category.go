package category

import "time"

type Category struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateCategoryInput struct {
	UserID int64  `json:"-"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}
