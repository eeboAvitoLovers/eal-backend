// Package model содержит определения структур данных для моделей обращений и пользователей.
package model

import (
	"database/sql"
	"time"
)

// Message представляет модель сообщения с полями для текста сообщения, идентификатора пользователя,
// времени создания и времени обновления.
type Message struct {
	ID         int    `json:"id,omitempty"`
	Message    string `json:"message"`
	UserID     int    `json:"user_id"`
	CreateAt   string `json:"create_at"`
	UpdateAt   string `json:"update_at"`
	Solved     string `json:"solved,omitempty"`
	Result     string `json:"result,omitempty"`
	ResolverID int    `json:"resolver_id"`
}

// MessageDTO представляет модель сообщения для передачи данных на клиент.
// Включает в себя идентификатор сообщения, текст сообщения, идентификатор пользователя,
// время создания, время обновления и флаг решения сообщения.
type MessageDTO struct {
	ID         int            `db:"id" json:"id"`
	UserID     int            `db:"user_id" json:"user_id"`
	UpdateAt   time.Time      `db:"update_at" json:"update_at"`
	CreateAt   time.Time      `db:"create_at" json:"create_at"`
	Message    string         `db:"message" json:"message"`
	Solved     sql.NullString `db:"solved" json:"solved"`
	Result     sql.NullString `db:"result" json:"result"`
	ResolverID sql.NullInt64  `db:"resolver_id" json:"resolver_id"`
}

type MessageValidDTO struct {
	ID         int       `db:"id" json:"id"`
	UserID     int       `db:"user_id" json:"user_id"`
	UpdateAt   time.Time `db:"update_at" json:"update_at"`
	CreateAt   time.Time `db:"create_at" json:"create_at"`
	Message    string    `db:"message" json:"message"`
	Solved     string    `db:"solved" json:"solved"`
	Result     string    `db:"result" json:"result"`
	ResolverID int       `db:"resolver_id" json:"resolver_id"`
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

type GetTicketListStruct struct {
	Messages []MessageValidDTO `json:"messages"`
	Total    int               `json:"total"`
}

func Validate(message MessageDTO) MessageValidDTO {
	var res MessageValidDTO
	res.ID = message.ID
	res.Message = message.Message
	res.UpdateAt = message.UpdateAt
	res.CreateAt = message.CreateAt
	res.UserID = message.UserID
	if !message.Solved.Valid {
		res.Solved = ""
	} else {
		res.Solved = message.Solved.String
	}
	if !message.Result.Valid {
		res.Result = ""
	} else {
		res.Result = message.Result.String
	}
	if message.ResolverID.Valid {
		res.ResolverID = int(message.ResolverID.Int64)
	} else {
		res.ResolverID = 0
	}
	return res
}

type Metric1 struct {
	Date    []time.Time `json:"date"`
	Percent []int       `json:"percent"`
}

type Metric2 struct {
	Cluster string `json:"cluster"`
	Topic   string `json:"topic"`
	Count   int    `json:"count"`
}
