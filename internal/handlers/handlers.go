// Package handlers содержит обработчики HTTP-запросов для взаимодействия с базой данных и моделями данных.
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/eeboAvitoLovers/eal-backend/internal/database"
	"github.com/eeboAvitoLovers/eal-backend/internal/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// MessageController предоставляет обработчики для управления сообщениями.
type MessageController struct {
	Controller *database.Controller
}

// CreateUserHandler обрабатывает запрос на создание нового пользователя.
// Принимает HTTP-запрос и записывает данные о новом пользователе в базу данных.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	var user model.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Print("error decoding")
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}

	userID, err := c.Controller.CreateUser(r.Context(), user, hashedPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}

	userResponse := model.UserDTO{
		ID:         userID,
		Email:      user.Email,
		IsEngineer: user.IsEngineer,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(userResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}
}

// LoginHandler обрабатывает запрос на аутентификацию пользователя.
// Принимает HTTP-запрос, аутентифицирует пользователя и создает новую сессию.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	var user model.UserLogin
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}

	ph, err := c.Controller.GetHash(r.Context(), user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}
	err = bcrypt.CompareHashAndPassword([]byte(ph), []byte(user.Password))
	if err != nil {
		http.Error(w, "password or email is incorrect", http.StatusUnauthorized)
		
	}
	sessionID := uuid.New().String()
	currentTime := time.Now()
	expAt := currentTime.Add(60 * time.Minute)
	err = c.Controller.CreateSession(r.Context(), user.Email, sessionID, expAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}

	cookie := http.Cookie{
		Name:    "session_id",
		Value:   sessionID,
		Expires: expAt,
		Path:    "/",
	}

	http.SetCookie(w, &cookie)

	var userResponse model.UserDTO
	isEngineer, err := c.Controller.IsEngineer(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}

	userID, err := c.Controller.GetUserIDBySessionID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}
	userResponse = model.UserDTO{
		ID:         userID,
		Email:      user.Email,
		IsEngineer: isEngineer,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(&userResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateMessage создает новое сообщение в базе данных.
// Принимает HTTP-запрос и данные нового сообщения.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) CreateMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	_, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		
	}

	var requestBody map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}

	message, ok := requestBody["message"].(string)
	if !ok {
		http.Error(w, "invalid JSON structure: message field is missing or not a string", http.StatusBadRequest)
		
	}

	sessionCookie, _ := r.Cookie("session_id")
	sessionID := sessionCookie.Value
	userID, err := c.Controller.GetUserIDBySessionID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	messageData := model.Message{
		Message:    message,
		UserID:     userID,
		CreateAt:   time.Now().Format("2006-01-02 15:04:05"),
		UpdateAt:   time.Now().Format("2006-01-02 15:04:05"),
		Solved:     "in_queue",
		ResolverID: 0,
	}

	messageID, err := c.Controller.CreateMessage(r.Context(), messageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}

	responseData := map[string]int{"id": messageID}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(&responseData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}

}

// GetStatusByID возвращает информацию о сообщении по его идентификатору.
// Принимает HTTP-запрос и идентификатор сообщения.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
// TODO заменить мапу
func (c *MessageController) GetStatusByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	_, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}

	message, err := c.Controller.GetStatusByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}
	w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}
}

// UserHasAcess проверяет, есть ли у пользователя доступ к системе.
// Принимает HTTP-запрос и возвращает true, если у пользователя есть доступ, и ошибку в случае отсутствия доступа.
func (c *MessageController) UserHasAcess(r *http.Request) (bool, error) {
	ctx := r.Context()
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		return false, fmt.Errorf("not authorized: %w", err)
	}
	sessionID := sessionCookie.Value

	isEngineer, err := c.Controller.IsEngineer(ctx, sessionID)
	return isEngineer, err
}

func (c *MessageController) MeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	_, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	sessionCookie, _ := r.Cookie("session_id")
	sessionID := sessionCookie.Value

	userID, err := c.Controller.GetUserIDBySessionID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}

	user, err := c.Controller.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}

}

func (c *MessageController) GetTicketList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	isEngineer, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		
	}
	if !isEngineer {
		http.Error(w, "no rights", http.StatusForbidden)
		
	}

	status := r.URL.Query().Get("status")
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}
	log.Print(status, offset, limit)

	tickets, err := c.Controller.GetTicketList(r.Context(), status, offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(&tickets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *MessageController) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	_, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		
	}

	sessionCookie, _ := r.Cookie("session_id")
	err = c.Controller.DeleteSession(r.Context(), sessionCookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}
	sessionCookie.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, sessionCookie)
}

func (c *MessageController) GetUnsolvedTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	isEngineer, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}
	if !isEngineer {
		http.Error(w, "no rights", http.StatusForbidden)
	}

	type requestBody struct {
		TicketID int `json:"ticket_id"`
	}

	var ticket requestBody

	err = json.NewDecoder(r.Body).Decode(&ticket)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	ticketID := ticket.TicketID

	vars := mux.Vars(r)
	resolverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	message, err := c.Controller.GetUnsolvedTicket(r.Context(), ticketID, resolverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *MessageController) UpdateStatusInProcess(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	isEngineer, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		
	}
	if !isEngineer {
		http.Error(w, "no rigths", http.StatusForbidden)
		
	}

	sessionCookie, _ := r.Cookie("session_id")
	sessionID := sessionCookie.Value

	userID, err := c.Controller.GetUserIDBySessionID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}

	resolverID, err := c.Controller.GetResolverIDByTicketID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}

	if resolverID != userID {
		http.Error(w, "resolverID != userID", http.StatusForbidden)
		
	}

	type statusResult struct {
		Status string `json:"status"`
		Result string `json:"result,omitempty"`
	}

	var statusStr statusResult
	err = json.NewDecoder(r.Body).Decode(&statusStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}

	log.Print("change ticket status", "id", id, "userID", userID, "status", statusStr)
	message, err := c.Controller.UpdateStatusInProgress(r.Context(), id, userID, statusStr.Status, statusStr.Result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}
}

func (c *MessageController) GetMyTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	isEngineer, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		
	}
	if !isEngineer {
		http.Error(w, "no rigths", http.StatusForbidden)
		
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		
	}
	log.Print("get my tickets func", offset, limit)

	sessionCookie, _ := r.Cookie("session_id")
	sessionID := sessionCookie.Value
	resolverID, err := c.Controller.GetUserIDBySessionID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}
	log.Print(resolverID)
	response, err := c.Controller.GetMyTickets(r.Context(), limit, offset, resolverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Print(response)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		
	}
}

func (c *MessageController) Analytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	_, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		
	}

	type AVGTime struct {
		AiP time.Duration `json:"accepted_in_progress"`
		AS  time.Duration `json:"accepted_solved"`
	}
	type ClosedTickets struct {
		Total     int `json:"total"`
		ThisMonth int `json:"this_month"`
		PrevMonth int `json:"prev_month"`
	}
	type Response struct {
		AVG     AVGTime         `json:"avg_time"`
		Closed  ClosedTickets   `json:"closed_tickets"`
		Metric1 model.Metric1   `json:"metric1"`
		Metric2 []model.Metric2 `json:"metric2"`
	}
	now := time.Hour
	avg := AVGTime{
		AiP: now,
		AS:  now,
	}

	metric1, err := c.Controller.GetMetric1(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	metric2, err := c.Controller.GetMetric2(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	thisMonth, err := c.Controller.AnalyticsThisMonth(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	closed := ClosedTickets{
		Total:     57,
		ThisMonth: thisMonth,
		PrevMonth: 97,
	}
	avgTime := Response{
		AVG:     avg,
		Closed:  closed,
		Metric1: metric1,
		Metric2: metric2,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(avgTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
