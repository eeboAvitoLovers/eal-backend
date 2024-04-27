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
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081, https://eal-frontend.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	var user model.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Print("error decoding")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	userID, err := c.Controller.CreateUser(r.Context(), user, hashedPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		return
	}
}

// LoginHandler обрабатывает запрос на аутентификацию пользователя.
// Принимает HTTP-запрос, аутентифицирует пользователя и создает новую сессию.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081, https://eal-frontend.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

	var user model.UserLogin
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ph, err := c.Controller.GetHash(r.Context(), user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(ph), []byte(user.Password))
	if err != nil {
		http.Error(w, "password or email is incorrect", http.StatusUnauthorized)
		return
	}
	sessionID := uuid.New().String()
	currentTime := time.Now()
	expAt := currentTime.Add(60 * time.Minute)
	err = c.Controller.CreateSession(r.Context(), user.Email, sessionID, expAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		return
	}

	userID, err := c.Controller.GetUserIDBySessionID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

// GetUnsolved возвращает нерешенные сообщения из базы данных.
// Принимает HTTP-запрос и возвращает нерешенные сообщения в формате JSON.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) GetUnsolved(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081, https://eal-frontend.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	isEngineer, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !isEngineer {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var messages []model.MessageDTO
	messages, err = c.Controller.GetUnsolved(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(&messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetUnsolvedID возвращает нерешенное сообщение по его идентификатору из базы данных.
// Принимает HTTP-запрос и идентификатор сообщения.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) GetUnsolvedID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081, https://eal-frontend.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	_, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	message, err := c.Controller.GetUnsolvedID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// UpdateSolvedID обновляет статус решения сообщения в базе данных по его идентификатору.
// Принимает HTTP-запрос и идентификатор сообщения.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) UpdateSolvedID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081, https://eal-frontend.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	isEngineer, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !isEngineer {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	timeNow := time.Now()
	messageResponse, err := c.Controller.UpdateSolvedID(r.Context(), timeNow, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(&messageResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// CreateMessage создает новое сообщение в базе данных.
// Принимает HTTP-запрос и данные нового сообщения.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) CreateMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081, https://eal-frontend.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	_, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var requestBody map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, ok := requestBody["message"].(string)
    if !ok {
        http.Error(w, "invalid JSON structure: message field is missing or not a string", http.StatusBadRequest)
        return
    }

	sessionCookie, _ := r.Cookie("session_id")
	sessionID := sessionCookie.Value
	userID, err := c.Controller.GetUserIDBySessionID(r.Context(), sessionID)
	messageData := model.Message{
		Message:  message,
		UserID:   userID,
		CreateAt: time.Now().Format("2006-01-02 15:04:05"),
		UpdateAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	messageID, err := c.Controller.CreateMessage(r.Context(), messageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseData := map[string]int{"id": messageID}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(&responseData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// GetStatusByID возвращает информацию о сообщении по его идентификатору.
// Принимает HTTP-запрос и идентификатор сообщения.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
// TODO заменить мапу
func (c *MessageController) GetStatusByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081, https://eal-frontend.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	m := map[bool]string{
		true: "solved",
		false: "accepted",
	}

	_, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, err := c.Controller.GetStatusByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	messageResponse := model.MessageResponse{
		ID: message.ID,      
		Message:  message.Message,
		UserID:   message.UserID,
		CreateAt: message.CreateAt,
		UpdateAt: message.UpdateAt,
		Solved:   m[message.Solved],
	}
	w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(&messageResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081, https://eal-frontend.vercel.app")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	_, err := c.UserHasAcess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	sessionCookie, _ := r.Cookie("session_id")
	sessionID := sessionCookie.Value

	userID, err := c.Controller.GetUserIDBySessionID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := c.Controller.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
