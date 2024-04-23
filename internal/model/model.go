package model

type Message struct {
	Message    string `json:"message"`
	UserID     int    `json:"user_id"`
	CreateAt   int    `json:"create_at"`
	UpdateAt   int    `json:"update_at"`
	RootID     int    `json:"root_id"`
	EditAt     int    `json:"edit_at"`
	DeleteAt   int    `json:"delete_at"`
	IsPinned   bool   `json:"is_pinned"`
	OriginalID int    `json:"original_id"`
}

type MessageDTO struct {
	ID         int    `json:"id"`
	Message    string `json:"message"`
	UserID     int    `json:"user_id"`
	CreateAt   int    `json:"create_at"`
	UpdateAt   int    `json:"update_at"`
	RootID     int    `json:"root_id"`
	EditAt     int    `json:"edit_at"`
	DeleteAt   int    `json:"delete_at"`
	IsPinned   bool   `json:"is_pinned"`
	OriginalID int    `json:"original_id"`
}
