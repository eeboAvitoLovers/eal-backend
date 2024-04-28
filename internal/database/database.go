// Package database предоставляет функции для взаимодействия с базой данных.
// Эти функции используются для выполнения операций CRUD (создание, чтение, обновление, удаление)
// на моделях данных, таких как пользователи и сообщения.

package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

// CreateMessage создает новое сообщение в базе данных.
// Принимает контекст и данные нового сообщения.
// Возвращает ошибку, если создание не удалось.
func (c *Controller) CreateMessage(ctx context.Context, message model.Message) (int, error) {
	messageID, err := c.GetNewID(ctx)
	if err != nil {
		return 0, fmt.Errorf("unable to get new id: %w", err)
	}

	// clusterID, err := c.GiveClusterID(ctx, messageID, message.Message)
	// if err != nil {
	// 	return 0, fmt.Errorf("cannot get clusterID: %w", err)
	// }

	query := `
        INSERT INTO messages (id, message, user_id, create_at, update_at, solved)
        VALUES ($1, $2, $3, $4, $5, $6);
    `

	// Выполняем SQL-запрос и сканируем результаты в переменную messageID.
	_, err = c.Client.Exec(ctx, query, messageID, message.Message, message.UserID, message.CreateAt, message.UpdateAt, message.Solved)
	if err != nil {
		return 0, fmt.Errorf("unable to create message: %w", err)
	}

	// query = `
	// 	INSERT INTO clusters (ticket_id, cluster) 
	// 	VALUES ($1, $2)
	// `
	// _, err = c.Client.Exec(ctx, query, messageID, clusterID)
	// if err != nil {
	// 	return 0, fmt.Errorf("unable to insert into clusters: %w", err)
	// }
	return int(messageID), nil
}

// GetStatusByID возвращает информацию о сообщении по его идентификатору.
// Принимает контекст и идентификатор сообщения.
// Возвращает информацию о сообщении и ошибку, если сообщение не найдено или произошла ошибка.
func (c *Controller) GetStatusByID(ctx context.Context, messageID int) (model.MessageValidDTO, error) {
	var message model.MessageDTO
	var resolverID sql.NullInt64
	query := `
	SELECT id, user_id, update_at, create_at, message, solved, result, resolver_id
	FROM (
		SELECT *, ROW_NUMBER() OVER (PARTITION BY id ORDER BY update_at DESC) AS rn
		FROM messages
		WHERE id = $1
	) AS CTE
	WHERE rn = 1
	`
	err := c.Client.QueryRow(ctx, query, messageID).
    Scan(&message.ID, &message.UserID, &message.UpdateAt, &message.CreateAt, &message.Message, &message.Solved, &message.Result, &resolverID)

	valid := model.Validate(message)

	if err != nil {
		return model.MessageValidDTO{}, err
	}
	return valid, nil
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

func (c *Controller) GetTicketList(ctx context.Context, status string, offset, limit int) (model.GetTicketListStruct, error) {
	query := `
		SELECT id, user_id, update_at, create_at, message, solved, result, resolver_id
		FROM (
			SELECT *, ROW_NUMBER() OVER (PARTITION BY id ORDER BY update_at DESC) AS rn
			FROM messages
			WHERE solved = $1
		) AS CTE
		WHERE rn = 1
		LIMIT $2 OFFSET $3
	`
	var messages []model.MessageValidDTO
	rows, err := c.Client.Query(context.Background(), query, status, limit, offset)
	if err != nil {
		return model.GetTicketListStruct{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var message model.MessageDTO
		if err := rows.Scan(&message.ID, &message.UserID, &message.UpdateAt, &message.CreateAt, &message.Message, &message.Solved, &message.Result, &message.ResolverID); err != nil {
			return model.GetTicketListStruct{}, err
		}
		
		messages = append(messages, model.Validate(message))
	}
	if err := rows.Err(); err != nil {
		return model.GetTicketListStruct{}, err
	}

	var cnt int
	query = `
		SELECT COUNT(id)
		FROM (
			SELECT *, ROW_NUMBER() OVER (PARTITION BY id ORDER BY update_at DESC) AS rn
			FROM messages
			WHERE solved = $1
		) AS CTE
		WHERE rn = 1
	`
	conn, err := c.Client.Acquire(ctx)
	if err != nil {
		return model.GetTicketListStruct{}, err
	}
	defer conn.Release()

	row := conn.QueryRow(ctx, query, status)
	err = row.Scan(&cnt)
	if err != nil {
		return model.GetTicketListStruct{}, err
	}

	response := model.GetTicketListStruct{
		Messages: messages,
		Total:    cnt,
	}
	log.Print(response)

	return response, nil
}

func (c *Controller) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE session_id=$1`

	_, err := c.Client.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}

	return nil
}

func (c *Controller) UpdateStatusInProgress(ctx context.Context, ticketID, resolverID int, status, result string) (model.MessageDTO, error) {
	messageID, err := c.GetNewID(ctx)
	if err != nil {
		return model.MessageDTO{}, fmt.Errorf("unable to get new id: %w", err)
	}

	var message model.MessageDTO
	err = c.Client.QueryRow(ctx, "SELECT id, user_id, update_at, create_at, message, solved, result, resolver_id FROM messages WHERE id = $1", ticketID).
		Scan(&message.ID, &message.UserID, &message.UpdateAt, &message.CreateAt, &message.Message, &message.Solved, &message.Result, &message.ResolverID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return model.MessageDTO{}, fmt.Errorf("no messages with provided id")
		} else {
			return model.MessageDTO{}, fmt.Errorf("unable to get message: %w", err)
		}
	}

	message.UpdateAt = time.Now()
	updateAtStr := message.UpdateAt.Format("2006-01-02 15:04:05")

	// Подготовка запроса на вставку данных
	var query string
	var args []interface{}

	if result == "" {
		query = `
			INSERT INTO messages (id, message, user_id, create_at, update_at, solved, resolver_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		args = []interface{}{messageID, message.Message, message.UserID, message.CreateAt, updateAtStr, message.Solved, resolverID}
	} else {
		query = `
			INSERT INTO messages (id, message, user_id, create_at, update_at, solved, resolver_id, result)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		args = []interface{}{messageID, message.Message, message.UserID, message.CreateAt, updateAtStr, message.Solved, resolverID, result}
	}

	// Выполнение запроса на вставку
	_, err = c.Client.Exec(ctx, query, args...)
	if err != nil {
		return model.MessageDTO{}, fmt.Errorf("error inserting message: %w", err)
	}

	return message, nil
}


func (c *Controller) GetUnsolvedTicket(ctx context.Context, ticketID, resolverID int) (model.MessageValidDTO, error) {
	status := "in_progress"
	updateAt := time.Now().Format("2006-01-02 15:04:05")
	var resolver sql.NullInt64
	err := c.Client.QueryRow(ctx, "SELECT resolver_id FROM messages WHERE id=$1", ticketID).Scan(&resolver)
	if err != nil {
		return model.MessageValidDTO{}, err
	}
	newTicketID, err := c.GetNewID(ctx)
	if err != nil {
		return model.MessageValidDTO{}, err
	}

	oldMessage, err := c.GetTicketByID(ctx, ticketID)
	if err != nil {
		return model.MessageValidDTO{}, err
	}

	log.Print(status)
	if !resolver.Valid {
		_, err = c.Client.Exec(ctx, `
        INSERT INTO messages (id, message, user_id, create_at, update_at, solved, resolver_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7);
    `, newTicketID, oldMessage.Message, oldMessage.UserID,oldMessage.CreateAt, updateAt, status, resolverID)
		if err != nil {
			return model.MessageValidDTO{}, fmt.Errorf("unable to update status: %w", err)
		}
	}

	ticket, err := c.GetTicketByID(ctx, newTicketID)
	if err != nil {
		return model.MessageValidDTO{}, fmt.Errorf("unable to get ticket: %w", err)
	}
	res := model.Validate(ticket)


	return res, nil
}

func (c *Controller) GetMyTickets(ctx context.Context, limit, offset, userID int) (model.GetTicketListStruct, error) {
    messages := make([]model.MessageValidDTO, 0, limit)
    
    // SQL query to retrieve ticket data
    query := `
        SELECT id, user_id, update_at, create_at, message, solved, result, resolver_id
        FROM messages
        WHERE resolver_id = $1
        ORDER BY update_at DESC
        LIMIT $2 OFFSET $3;
    `
    
    rows, err := c.Client.Query(ctx, query, userID, limit, offset)
    if err != nil {
        return model.GetTicketListStruct{}, err
    }
    defer rows.Close()
	
    for rows.Next() {
        var message model.MessageDTO
        err := rows.Scan(&message.ID, &message.UserID, &message.UpdateAt, &message.CreateAt, &message.Message, &message.Solved, &message.Result, &message.ResolverID)
        if err != nil {
            return model.GetTicketListStruct{}, err
        }
		log.Print("db mytickets func", model.Validate(message))
        messages = append(messages, model.Validate(message))
    }

    var cnt int
    // SQL query to count total number of tickets
    countQuery := `
        SELECT COUNT(id)
        FROM messages
        WHERE resolver_id = $1;
    `
    conn, err := c.Client.Acquire(ctx)
    if err != nil {
        return model.GetTicketListStruct{}, err
    }
    defer conn.Release()

    row := conn.QueryRow(ctx, countQuery, userID)
    err = row.Scan(&cnt) // Scan into integer variable directly
    if err != nil {
        return model.GetTicketListStruct{}, err
    }

    return model.GetTicketListStruct{
        Messages: messages,
        Total:    cnt,
    }, nil
}
