package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/eeboAvitoLovers/eal-backend/internal/model"
	"github.com/eeboAvitoLovers/eal-backend/internal/database"
	// "github.com/gorilla/mux"
)

type MessageController struct {
	Controller *database.Controller
}

func (c *MessageController) CreateMessageHandler(w http.ResponseWriter, r *http.Request) {
	var message model.Message
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = c.Controller.Create(r.Context(), message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	// Убрать после дебага
	log.Print(message)
	w.WriteHeader(http.StatusCreated)

}