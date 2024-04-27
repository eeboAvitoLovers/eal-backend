// Package model содержит определения структур данных для моделей обращений и пользователей.
package model

import "time"

// Message представляет модель сообщения с полями для текста сообщения, идентификатора пользователя,
// времени создания и времени обновления.
type Message struct {
	Message       string `json:"message"`
	UserID        int    `json:"user_id"`
	CreateAt      string `json:"create_at"`
	UpdateAt      string `json:"update_at"`
	Solved        string `json:"solved,omitempty"`
	ResolverID    int    `json:"resolver_id"`
	WorkStartDate string `json:"work_start_date"`
}

// MessageDTO представляет модель сообщения для передачи данных на клиент.
// Включает в себя идентификатор сообщения, текст сообщения, идентификатор пользователя,
// время создания, время обновления и флаг решения сообщения.
type MessageDTO struct {
	ID            int       `json:"id,omitempty"`
	Message       string    `json:"message"`
	UserID        int       `json:"user_id"`
	CreateAt      time.Time `json:"create_at"`
	UpdateAt      time.Time `json:"update_at"`
	Solved        string    `json:"solved,omitempty"`
	ResolverID    int       `json:"resolver_id"`
	WorkStartDate time.Time `json:"work_start_date"`
}

// User представляет модель пользователя с полями для электронной почты, пароля и флага инженера.
type User struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	IsEngineer bool   `json:"is_engineer"`
}

type UserDTO struct {
	ID         int    `json:"ID"`
	Email      string `json:"email"`
	IsEngineer bool   `json:"is_engineer"`
}

// UserLogin представляет модель для аутентификации пользователя с полями для электронной почты и пароля.
type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type MessageResponse struct {
	ID       int       `json:"id"`
	Message  string    `json:"message"`
	UserID   int       `json:"user_id"`
	CreateAt time.Time `json:"create_at"`
	UpdateAt time.Time `json:"update_at"`
	Solved   string    `json:"solved"`
}

type GetTicketListStruct struct {
	Messages []MessageDTO `json:"messages"`
	Total    int          `json:"total"`
}

type AvgTime struct {
	AcceptedInProgress time.Time `json:"accepted_in_progress"`
	AcceptedSolved     time.Time `json:"accepted_solved"`
	InProgressSolved   time.Time `json:"in_progress_solved"`
}
