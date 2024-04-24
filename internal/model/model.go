// Package model содержит определения структур данных для моделей обращений и пользователей.
package model

import "time"

// Message представляет модель сообщения с полями для текста сообщения, идентификатора пользователя,
// времени создания и времени обновления.
type Message struct {
	Message  string `json:"message"`
	UserID   int    `json:"user_id"`
	CreateAt string `json:"create_at"`
	UpdateAt string `json:"update_at"`
}

// MessageDTO представляет модель сообщения для передачи данных на клиент.
// Включает в себя идентификатор сообщения, текст сообщения, идентификатор пользователя,
// время создания, время обновления и флаг решения сообщения.
type MessageDTO struct {
	ID       int       `json:"id"`
	Message  string    `json:"message"`
	UserID   int       `json:"user_id"`
	CreateAt time.Time `json:"create_at"`
	UpdateAt time.Time `json:"update_at"`
	Solved   bool      `json:"solved,omitempty"`
}

// User представляет модель пользователя с полями для электронной почты, пароля и флага инженера.
type User struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	IsEngineer bool   `json:"is_engineer"`
}

// UserLogin представляет модель для аутентификации пользователя с полями для электронной почты и пароля.
type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
