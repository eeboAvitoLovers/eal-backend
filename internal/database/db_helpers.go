package database

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/eeboAvitoLovers/eal-backend/internal/model"
	"github.com/eeboAvitoLovers/eal-backend/internal/config"
)

const configFilename = "./internal/config/config.yaml"

func (c *Controller) GetNewID(ctx context.Context) (int, error) {
	query := `
		SELECT 
		CASE 
		WHEN EXISTS (SELECT * FROM messages) THEN MAX(id) + 1
		ELSE 1
		END AS ID
		from messages`

	var messageID int64
	err := c.Client.QueryRow(ctx, query).Scan(&messageID)
	if err != nil {
		return 0, fmt.Errorf("unable to get id: %w", err)
	}
	return int(messageID), nil
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

func (c *Controller) GetResolverIDByTicketID(ctx context.Context, ticketID int) (int, error) {
	query := `SELECT resolver_id FROM messages WHERE id = $1`

	var resolverID sql.NullInt64

	row := c.Client.QueryRow(ctx, query, ticketID)
    err := row.Scan(&resolverID)
	if !resolverID.Valid {
		return 0, fmt.Errorf("resolverID is null")
	}
    if err != nil {
        if err.Error() == "no rows in result set" {
            return 0, fmt.Errorf("no message found with ID: %w", err)
        } else {
			return 0, fmt.Errorf("unable to get resolverID: %w", err)
        }
    }

	return int(resolverID.Int64), nil
}

func (c *Controller) GetTicketByID(ctx context.Context, ticketID int) (model.MessageDTO, error) {
	query := `
		SELECT id, user_id, update_at, create_at, message, solved, result, resolver_id
		FROM public.messages
		WHERE id = $1
	`
	var message model.MessageDTO

	err := c.Client.QueryRow(ctx, query, ticketID).Scan(
		&message.ID, &message.UserID, &message.UpdateAt, &message.CreateAt, &message.Message, &message.Solved, &message.Result, &message.ResolverID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return model.MessageDTO{}, fmt.Errorf("no message found with ID: %w", err)
		} else {
			return model.MessageDTO{}, fmt.Errorf("unable to get message: %w", err)
		}
	}
	return message, nil
}

func (c *Controller) GetNewTicket(ctx context.Context, newTicketID int) (model.MessageDTO, error) {
	query := `
		SELECT id, message, user_id, create_at, update_at, solved, resolver_id, result
		FROM messages
		WHERE id = $1
	`
	var message model.MessageDTO
	err := c.Client.QueryRow(ctx, query, newTicketID).Scan(
		&message.ID, &message.UserID, &message.UpdateAt, &message.CreateAt, &message.Message, &message.Solved, &message.Result, &message.ResolverID)
	if err != nil {
		return model.MessageDTO{}, err
	}
	return message, nil
}

func (c *Controller) GiveClusterID(ctx context.Context, newTicketID int, message string) (int, error) {
	data := map[string]interface{}{
		"message": message,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("unable to send request: %w", err)
	}

	reqBody := bytes.NewBuffer(jsonData)

	config, err := config.LoadConfig(configFilename)
	if err != nil {
		return 0, fmt.Errorf("unable to load config: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s:%d", config.Clusters.Hostname, config.Clusters.Port), "application/json", reqBody)
	if err != nil {
		return 0, fmt.Errorf("unable to send post request: %w", err)
	}

	type RespStruct struct {
		Message string `json:"message"`
		Cluster string `json:"cluster"`
	}

	var r RespStruct

	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return 0, fmt.Errorf("unable to decode response: %w", err)
	}

	clusterID, err := strconv.Atoi(r.Cluster)
	if err != nil {
		return 0, fmt.Errorf("unable to convert clusterID: %w", err)
	}

	return clusterID, nil
}