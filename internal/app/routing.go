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

// loadRoutes загружает маршруты в приложение.
// Принимает указатель на маршрутизатор mux.Router.
func (a *App) loadRoutes(r *mux.Router) {
	// Создание обработчика URL.
	urlHandler := &handlers.MessageController{
		// Создание экземпляра контроллера сообщений, который включает в себя экземпляр контроллера базы данных.
		Controller: &database.Controller{
			// Передача пула подключений к базе данных из приложения в контроллер базы данных.
			Client: a.pgpool,
		},
	}
	
	r.HandleFunc("/me/", urlHandler.MeHandler).Methods("GET")
	// POST /login - аутентификация пользователя по электронной почте и паролю
    // POST /register - регистрация нового пользователя
    // Оба эндпоинта ожидают JSON с электронной почтой и паролем в качестве данных.
    // Регистрация перенаправляет пользователя на страницу /login для входа.
	// Пример JSON запроса
	// {
	// 	"email": "ovchark4@yandex.ru",
	// 	"password": "812749iasldf83"
	// }
	r.HandleFunc("/login/", urlHandler.LoginHandler).Methods("POST")
	r.HandleFunc("/register/", urlHandler.CreateUserHandler).Methods("POST")
	// r.HandleFunc("/logout", urlHandler.LogoutHandler).Methods("GET")

    // Обработчики для специалистов

    // POST /ticket - создает новый запрос.
    // Запрос должен содержать JSON с текстом сообщения.
    // Возвращает код состояния 201 Created.
	// Пример JSON запроса 
	// {
	// 	"message": "Привет сломался вывод средств"
	// }
	// response 201 Created
	r.HandleFunc("/ticket/", urlHandler.CreateMessage).Methods("POST")

    // GET /ticket/{id} - получение информации о запросе по его идентификатору.
    // Ответ в формате JSON.
	// Пример JSON ответа
	// {
	// 	"id": 1234,
	// 	"message": "Привет сломался возврат средств",
	// 	"user_id": 127400,
	// 	"create_at": 128754691200,
	// 	"update_at": 12849013290,
	// 	"solved": "solved"
	// }
	// response 200 OK
	r.HandleFunc("/ticket/{id}", urlHandler.GetStatusByID).Methods("GET")
	// обновляет статус тикета на указанный
	r.HandleFunc("/ticket/{id}", urlHandler.UpdateStatusInProcess).Methods("PUT")	
	// Выводит список сообщений с указанным статусом
	r.HandleFunc("/ticket/", urlHandler.GetTicketList).Queries("status", "{status}", "offset", "{offset}", "limit", "{limit}").Methods("GET")
	// Присваивает тикет инженеру
	r.HandleFunc("/specialist/{id}/tickets", urlHandler.GetUnsolvedTicket).Methods("POST")
	// Выводит список тикетов принадлежащих инженеру
	r.HandleFunc("/specialist/{id}/tickets", urlHandler.GetMyTickets).Queries("offset", "{offset}", "limit", "{limit}").Methods("GET")

	//TODO
	// r.HandleFunc("/tickets/analytics", urlHandler.Analytics).Methods("GET")
}