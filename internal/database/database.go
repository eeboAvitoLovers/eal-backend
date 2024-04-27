// Package database предоставляет функции для взаимодействия с базой данных.
// Эти функции используются для выполнения операций CRUD (создание, чтение, обновление, удаление)
// на моделях данных, таких как пользователи и сообщения.

package database

import (
	"context"
	"fmt"
	// "strings"
	"time"

	"github.com/eeboAvitoLovers/eal-backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Controller представляет контроллер базы данных, который обеспечивает доступ к базе данных.
type Controller struct {
	Client *pgxpool.Pool
}

// CreateUser создает нового пользователя в базе данных.
// Принимает контекст и данные нового пользователя.
// Возвращает ошибку, если создание не удалось.
func (c *Controller) CreateUser(ctx context.Context, data model.User, hp []byte) (int, error) {
	// SQL-запрос для вставки нового пользователя.
	// query := `INSERT INTO users (email, password, is_engineer) VALUES (@email, @password, @is_engineer) RETURNING  id;`
	// // Аргументы для передачи в SQL-запрос в виде именованных аргументов.
	// var userID int64
	// err := c.Client.QueryRow(ctx, query, data.Email, string(hp), data.IsEngineer).Scan(&userID)
	// if err != nil {
	// 	return 0, fmt.Errorf("unable to create user: %w", err)
	// }
	var userID int64
	err := c.Client.QueryRow(ctx, "INSERT INTO users (email, password, is_engineer) VALUES ($1, $2, $3) RETURNING  id;",
		data.Email, string(hp), data.IsEngineer).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("error adding user: %w", err)
	}

	return int(userID), nil
}

// GetHash возвращает хешированный пароль пользователя по его email.
// Принимает контекст и email пользователя.
// Возвращает хешированный пароль и ошибку, если такого пользователя нет или произошла ошибка.
func (c *Controller) GetHash(ctx context.Context, email string) (string, error) {
	var ph string
	// Получение хешированного пароля из базы данных по email пользователя.
	conn, err := c.Client.Acquire(ctx)
	if err != nil {
		return "", fmt.Errorf("error acquiring connection from pool: %w", err)
	}
	defer conn.Release()
	err = conn.QueryRow(ctx, "SELECT password FROM users WHERE email = $1", email).Scan(&ph)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("user with email %s not found", email)
		}
		return "", fmt.Errorf("error getting hash: %w", err)
	}

	return ph, nil
}

// CreateSession создает новую сессию пользователя в базе данных.
// Принимает контекст, email пользователя, идентификатор сессии и время истечения сессии.
// Возвращает ошибку, если создание не удалось.
func (c *Controller) CreateSession(ctx context.Context, email, sessionID string, expAt time.Time) error {
	// Получение идентификатора пользователя и информации об инженерном статусе пользователя по его email.
	var userID int
	err := c.Client.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		return fmt.Errorf("error getting user info: %w", err)
	}

	// Добавление записи о сессии в таблицу sessions.
	_, err = c.Client.Exec(ctx, "INSERT INTO sessions (session_id, user_id, exp_at) VALUES ($1, $2, $3)",
		sessionID, userID, expAt)
	if err != nil {
		return fmt.Errorf("error adding session: %w", err)
	}

	return nil
}

// IsEngineer проверяет, является ли пользователь инженером по его идентификатору сессии.
// Принимает контекст и идентификатор сессии.
// Возвращает true, если пользователь инженер, и ошибку в случае неудачи.
func (c *Controller) IsEngineer(ctx context.Context, sessionID string) (bool, error) {
	var isEngineer bool
	err := c.Client.QueryRow(ctx, `
        SELECT u.is_engineer
        FROM sessions s
        JOIN users u ON s.user_id = u.id
        WHERE s.session_id = $1`, sessionID).Scan(&isEngineer)
	if err != nil {
		return false, fmt.Errorf("error cheking rights: %w", err)
	}
	return isEngineer, nil
}

// GetUnsolved возвращает нерешенные сообщения из базы данных.
// Принимает контекст.
// Возвращает срез сообщений и ошибку в случае неудачи.
func (c *Controller) GetUnsolved(ctx context.Context) ([]model.MessageDTO, error) {
	// Запрос к базе данных для получения нерешенных сообщений, ограниченных до 10 штук и отсортированных по времени создания.
	rows, err := c.Client.Query(context.Background(), "SELECT id, message, user_id, create_at, update_at, solved FROM messages WHERE solved=false ORDER BY create_at LIMIT 10")
	if err != nil {
		err = fmt.Errorf("unable to execute query: %w", err)
		return []model.MessageDTO{}, err
	}
	defer rows.Close()

	// Инициализация среза для хранения сообщений.
	var messages []model.MessageDTO
	// Итерация по результатам запроса.
	for rows.Next() {
		var message model.MessageDTO
		// Сканирование строки результата в структуру сообщения.
		err := rows.Scan(&message.ID, &message.Message, &message.UserID, &message.CreateAt, &message.UpdateAt, &message.Solved)
		if err != nil {
			err = fmt.Errorf("unable to scan row: %w", err)
			return []model.MessageDTO{}, err
		}
		// Добавление сообщения в срез.
		messages = append(messages, message)
	}

	// Проверка наличия ошибок после итерации по результатам.
	if err := rows.Err(); err != nil {
		err = fmt.Errorf("error after iterating rows: %w", err)
		return []model.MessageDTO{}, err
	}

	return messages, nil
}

// GetUnsolvedID возвращает информацию о нерешенном сообщении по его идентификатору.
// Принимает контекст и идентификатор сообщения.
// Возвращает информацию о сообщении и ошибку, если сообщение не найдено или произошла ошибка.
func (c *Controller) GetUnsolvedID(ctx context.Context, messageID int) (model.MessageDTO, error) {
	var message model.MessageDTO
	// Запрос к базе данных для получения информации о сообщении по его идентификатору.
	err := c.Client.QueryRow(ctx, "SELECT id, message, user_id, create_at, update_at, solved FROM messages WHERE id = $1 AND solved=false", messageID).Scan(
		&message.ID, &message.Message, &message.UserID, &message.CreateAt, &message.UpdateAt, &message.Solved)

	if err.Error() == "no rows in result set" {
		return model.MessageDTO{}, fmt.Errorf("no message with provided id")
	} else if err != nil {
		return model.MessageDTO{}, fmt.Errorf("error querying row: %w", err)
	}

	return message, nil
}

// UpdateSolvedID обновляет статус решения сообщения в базе данных по его идентификатору.
// Принимает контекст, время обновления и идентификатор сообщения.
// Возвращает ошибку, если обновление не удалось.
func (c *Controller) UpdateSolvedID(ctx context.Context, updateAt time.Time, messageID int) (model.MessageDTO, error) {
	// Установка статуса решения сообщения в true.
	solved := true
	// Обновление записи о сообщении в базе данных.
	_, err := c.Client.Exec(ctx, "UPDATE messages SET solved = $1, update_at = $2 WHERE id = $3", solved, updateAt, messageID)
	if err != nil {
		return model.MessageDTO{}, fmt.Errorf("unable to update solved status: %w", err)
	}
	var message model.MessageDTO
	message, err = c.GetUnsolvedID(ctx, messageID)
	if err != nil {
		return model.MessageDTO{}, fmt.Errorf("unable to get message: %w", err)
	}

	return message, nil
}

// GetUserIDBySessionID возвращает идентификатор пользователя по его идентификатору сессии.
// Принимает контекст и идентификатор сессии.
// Возвращает идентификатор пользователя и ошибку, если сессия не найдена или произошла ошибка.
func (c *Controller) GetUserIDBySessionID(ctx context.Context, sessionID string) (int, error) {
	var userID int
	err := c.Client.QueryRow(ctx, "SELECT user_id FROM sessions WHERE session_id = $1", sessionID).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// CreateMessage создает новое сообщение в базе данных.
// Принимает контекст и данные нового сообщения.
// Возвращает ошибку, если создание не удалось.
func (c *Controller) CreateMessage(ctx context.Context, message model.Message) (int, error) {
	query := `
        INSERT INTO messages (message, user_id, create_at, update_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id;
    `

	// Выполняем SQL-запрос и сканируем результаты в переменную messageID.
	var messageID int64
	err := c.Client.QueryRow(ctx, query, message.Message, message.UserID, message.CreateAt, message.UpdateAt).
		Scan(&messageID)
	if err != nil {
		return 0, fmt.Errorf("unable to create message: %w", err)
	}
	return int(messageID), nil
}

// GetStatusByID возвращает информацию о сообщении по его идентификатору.
// Принимает контекст и идентификатор сообщения.
// Возвращает информацию о сообщении и ошибку, если сообщение не найдено или произошла ошибка.
func (c *Controller) GetStatusByID(ctx context.Context, messageID int) (model.MessageDTO, error) {
	var message model.MessageDTO
	err := c.Client.QueryRow(ctx, "SELECT id, message, user_id, create_at, update_at, solved FROM messages WHERE id = $1", messageID).
		Scan(&message.ID, &message.Message, &message.UserID, &message.CreateAt, &message.UpdateAt, &message.Solved)
	if err != nil {
		return model.MessageDTO{}, err
	}
	return message, nil
}

func (c *Controller) GetUserByID(ctx context.Context, userID int) (model.UserDTO, error) {
	query := `SELECT id, email, is_engineer FROM users WHERE id = $1;`

	// Используем QueryRow для выполнения запроса и сканирования результатов в структуру UserDTO.
	var user model.UserDTO
	err := c.Client.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Email, &user.IsEngineer)
	if err != nil {
		return model.UserDTO{}, fmt.Errorf("error fetching user: %w", err)
	}

	return user, nil
}

// TODO переделать когда исправим статусы
func (c *Controller) GetTicketList(ctx context.Context, status bool, offset, limit int) ([]model.MessageResponse, error) {
	response := make([]model.MessageResponse, 0, limit)
	rows, err := c.Client.Query(ctx, "SELECT * FROM tickets LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return []model.MessageResponse{}, err
	}
	defer rows.Close()

	m := map[bool]string{
		true: "solved",
		false: "accepted",
	}

	for rows.Next() {
		var message model.MessageResponse
		var status bool
		err := rows.Scan(&message.ID, &message.Message, &message.UserID, &message.CreateAt, &message.UpdateAt, &status)
		message.Solved = m[status]
		if err != nil {
			return []model.MessageResponse{}, err
		}
		response = append(response, message)
	}

	return response, nil
}
