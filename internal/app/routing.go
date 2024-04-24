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
	
	// POST /login - аутентификация пользователя по электронной почте и паролю
    // POST /register - регистрация нового пользователя
    // Оба эндпоинта ожидают JSON с электронной почтой и паролем в качестве данных.
    // Регистрация перенаправляет пользователя на страницу /login для входа.
	// Пример JSON запроса
	// {
	// 	"email": "ovchark4@yandex.ru",
	// 	"password": "812749iasldf83"
	// }
	r.HandleFunc("/login", urlHandler.LoginHandler).Methods("POST")
	r.HandleFunc("/register", urlHandler.CreateUserHandler).Methods("POST")

   	// Эндпоинт для перенаправления пользователей в зависимости от их роли
    // GET / - перенаправление пользователей на /engineer/ или /specialist/ в зависимости от их роли.
    // Также проверяется наличие у пользователя куки.
	r.HandleFunc("/", urlHandler.RedirectAccordingToRights).Methods("GET")

 	// Обработчики для инженеров

    // GET /engineer/ - получение списка нерешенных запросов.
    // Ответ в формате JSON.
	// Пример JSON ответа.
	// [{
	// 	"id": 1234,
	// 	"message": "Привет сломался возврат средств",
	// 	"user_id": 127400,
	// 	"create_at": 128754691200,
	// 	"update_at": 12849013290,
	// 	"solved": true
	// },
	// {
	// 	"id": 2134,
	// 	"message": "Привет сломались уведомления об остатке средств",
	// 	"user_id": 23556,
	// 	"create_at": 8938691200,
	// 	"update_at": 12847162491,
	// 	"solved": true
	// }]
	// 200 OK
	r.HandleFunc("/engineer/", urlHandler.GetUnsolved).Methods("GET")

   	// GET /engineer/{id} - получение одного нерешенного запроса по его идентификатору.
    // Ответ в формате JSON.
	// Пример JSON ответа.
	// {
	// 	"id": 1234,
	// 	"message": "Привет сломался возврат средств",
	// 	"user_id": 127400,
	// 	"create_at": 128754691200,
	// 	"update_at": 12849013290,
	// 	"solved": true
	// }
	// return 200 OK
	r.HandleFunc("/engineer/:id", urlHandler.GetUnsolvedID).Methods("GET")

    // PUT /engineer/{id} - помечает запрос с указанным идентификатором как решенный.
    // Возвращает код состояния 200 OK.
	r.HandleFunc("/engineer/:id", urlHandler.UpdateSolvedID).Methods("PUT")

    // Обработчики для специалистов

    // POST /specialist/ - создает новый запрос.
    // Запрос должен содержать JSON с текстом сообщения.
    // Возвращает код состояния 201 Created.
	// Пример JSON запроса 
	// {
	// 	"message": "Привет сломался вывод средств"
	// }
	// response 201 Created
	r.HandleFunc("/specialist/", urlHandler.CreateMessage).Methods("POST")

    // GET /specialist/{id} - получение информации о запросе по его идентификатору.
    // Ответ в формате JSON.
	// Пример JSON ответа
	// {
	// 	"id": 1234,
	// 	"message": "Привет сломался возврат средств",
	// 	"user_id": 127400,
	// 	"create_at": 128754691200,
	// 	"update_at": 12849013290,
	// 	"solved": true
	// }
	// response 200 OK
	r.HandleFunc("/specialist/:id", urlHandler.GetStatusByID).Methods("GET")
}