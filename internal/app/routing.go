package app

import (
	"github.com/eeboAvitoLovers/eal-backend/internal/database"
	"github.com/eeboAvitoLovers/eal-backend/internal/handlers"
	"github.com/gorilla/mux"
)

func (a *App) newRoutes() {
	r := mux.NewRouter()
	a.router = r

	a.loadRoutes(r)
}

func (a *App) loadRoutes(r *mux.Router) {
	urlHandler := &handlers.MessageController{
		Controller: &database.Controller{
			Client: a.pgpool,
		},
	}
	

	r.HandleFunc("/", handlers.CreateMessageHandler).Methods("POST")
	// r.HandleFunc("/archive", handlers.GetArchiveHandler).Methods("GET")
	// r.HandleFunc("/list", handlers.GetMessages).Methods("GET")
	// r.HandleFunc("/:id", handlers.GetMessageByID).Methods("GET")
}