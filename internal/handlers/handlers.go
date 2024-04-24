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
	var user model.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Print("error decoding")
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	err = c.Controller.CreateUser(r.Context(), user, hashedPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// LoginHandler обрабатывает запрос на аутентификацию пользователя.
// Принимает HTTP-запрос, аутентифицирует пользователя и создает новую сессию.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) LoginHandler(w http.ResponseWriter, r *http.Request) {
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
	http.Redirect(w, r, "/", http.StatusFound)
}

// RedirectAccordingToRights перенаправляет пользователя в зависимости от его прав.
// Принимает HTTP-запрос и в зависимости от статуса пользователя перенаправляет его на соответствующую страницу.
func (c *MessageController) RedirectAccordingToRights(w http.ResponseWriter, r *http.Request) {
	isEngineer, err := c.UserHasAcess(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
	}
	if isEngineer {
		http.Redirect(w, r, "/engineer/", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/specialist/", http.StatusSeeOther)
	}
}

// GetUnsolved возвращает нерешенные сообщения из базы данных.
// Принимает HTTP-запрос и возвращает нерешенные сообщения в формате JSON.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) GetUnsolved(w http.ResponseWriter, r *http.Request) {
	isEngineer, err := c.UserHasAcess(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
	}
	if !isEngineer {
		http.Redirect(w, r, "/", http.StatusForbidden)
	}

	var messages []model.MessageDTO
	messages, err = c.Controller.GetUnsolved(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetUnsolvedID возвращает нерешенное сообщение по его идентификатору из базы данных.
// Принимает HTTP-запрос и идентификатор сообщения.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) GetUnsolvedID(w http.ResponseWriter, r *http.Request) {
	isEngineer, err := c.UserHasAcess(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
	}
	if !isEngineer {
		http.Redirect(w, r, "/", http.StatusForbidden)
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	message, err := c.Controller.GetUnsolvedID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// UpdateSolvedID обновляет статус решения сообщения в базе данных по его идентификатору.
// Принимает HTTP-запрос и идентификатор сообщения.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) UpdateSolvedID(w http.ResponseWriter, r *http.Request) {
	isEngineer, err := c.UserHasAcess(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
	}
	if !isEngineer {
		http.Redirect(w, r, "/", http.StatusForbidden)
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	timeNow := time.Now()
	err = c.Controller.UpdateSolvedID(r.Context(), timeNow, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

// CreateMessage создает новое сообщение в базе данных.
// Принимает HTTP-запрос и данные нового сообщения.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) CreateMessage(w http.ResponseWriter, r *http.Request) {
	isEngineer, err := c.UserHasAcess(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
	}
	if !isEngineer {
		http.Redirect(w, r, "/", http.StatusForbidden)
	}

	var message string
	err = json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	err = c.Controller.CreateMessage(r.Context(), messageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

// GetStatusByID возвращает информацию о сообщении по его идентификатору.
// Принимает HTTP-запрос и идентификатор сообщения.
// В случае ошибки отправляет соответствующий HTTP-статус и сообщение об ошибке.
func (c *MessageController) GetStatusByID(w http.ResponseWriter, r *http.Request) {
	isEngineer, err := c.UserHasAcess(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
	}
	if !isEngineer {
		http.Redirect(w, r, "/", http.StatusForbidden)
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
	w.WriteHeader(http.StatusOK)
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
