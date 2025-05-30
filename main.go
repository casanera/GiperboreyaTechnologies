package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/casanera/GiperboreyaTechnologies/internal/handlers"
	"github.com/casanera/GiperboreyaTechnologies/internal/storage"
)

var db *sql.DB

// routeHandler будет диспетчеризировать запросы к UserHandler
func routeHandler(userH *handlers.UserHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("API Запрос: Метод=%s, Путь=%s", r.Method, r.URL.Path)

		// Только пути, начинающиеся с /api/v1/users
		if !strings.HasPrefix(r.URL.Path, "/api/v1/users") {
			log.Printf("Маршрут не API: %s, отдаем 404", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		// Определяем, есть ли ID в пути /api/v1/users/{id}
		// pathRemainder будет содержать ID или будет пустым/слешем
		pathRemainder := strings.TrimPrefix(r.URL.Path, "/api/v1/users")
		isSpecificUserPath := pathRemainder != "" && pathRemainder != "/"

		switch r.Method {
		case http.MethodGet:
			userH.GetUserHandler(w, r) // GetUserHandler сам разберет, всех или одного
		case http.MethodPost:
			// POST только на /api/v1/users (без ID)
			if !isSpecificUserPath {
				userH.CreateUserHandler(w, r)
			} else {
				http.Error(w, "Метод POST применим только к /api/v1/users", http.StatusMethodNotAllowed)
			}
		case http.MethodPut:
			// PUT только на /api/v1/users/{id}
			if isSpecificUserPath {
				userH.UpdateUserHandler(w, r)
			} else {
				http.Error(w, "Для PUT запроса требуется ID пользователя в пути (например, /api/v1/users/123)", http.StatusBadRequest)
			}
		case http.MethodDelete:
			// DELETE только на /api/v1/users/{id}
			if isSpecificUserPath {
				userH.DeleteUserHandler(w, r)
			} else {
				http.Error(w, "Для DELETE запроса требуется ID пользователя в пути (например, /api/v1/users/123)", http.StatusBadRequest)
			}
		default:
			http.Error(w, "Метод не разрешен для данного пути", http.StatusMethodNotAllowed)
		}
	}
}

func main() {
	log.Println("Запуск backend приложения с CRUD...")

	dbHost := os.Getenv("DB_HOST") // Для Docker Compose это будет имя сервиса 'db'
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Проверка обязательных переменных окружения
	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatalf("Одна или несколько переменных окружения для БД не установлены.")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Println("Попытка подключения к PostgreSQL...")
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка при вызове sql.Open для PostgreSQL: %v", err)
	}

	maxRetries := 15
	for i := 0; i < maxRetries; i++ {
		log.Printf("Проверка соединения с БД (попытка %d/%d)...", i+1, maxRetries)
		err = db.Ping()
		if err == nil {
			log.Println("Успешное подключение к PostgreSQL!")
			break
		}
		log.Printf("Не удалось подключиться к БД: %v. Ожидание 5 секунд перед следующей попыткой...", err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Не удалось установить соединение с PostgreSQL после %d попыток: %v. Завершение работы.", maxRetries, err)
	}

	// Инициализация хранилища
	userStore := storage.NewPostgresUserStorage(db)

	// Попытка создать таблицу users, если она не существует
	if pgs, ok := userStore.(*storage.PostgresUserStorage); ok {
		if err := pgs.CreateUsersTableIfNotExists(); err != nil {
			log.Fatalf("Не удалось создать/проверить таблицу пользователей: %v", err)
		}
	} else {
		log.Println("Предупреждение: не удалось привести userStore к *storage.PostgresUserStorage для создания таблицы.")
	}

	// Инициализация обработчика
	userHandler := handlers.NewUserHandler(userStore)

	// Настройка маршрутизатора
	mux := http.NewServeMux()

	// API маршруты
	mux.HandleFunc("/api/v1/users/", routeHandler(userHandler)) // routeHandler уже есть выше

	// Раздача статических файлов для всех остальных путей
	// Создаем обработчик для статических файлов из папки "static"
	staticFileServer := http.FileServer(http.Dir("./static"))
	// Все запросы, не начинающиеся с /api/, будут обработаны staticFileServer
	// Если запрошен "/", FileServer автоматически попытается отдать "index.html" из "./static"
	mux.Handle("/", staticFileServer) 

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	log.Printf("Сервер (с фронтендом) запускается на http://localhost:%s", appPort)
	log.Printf("API пользователей доступно по /api/v1/users")
	log.Printf("Фронтенд доступен по адресу: http://localhost:%s/", appPort)


	if err := http.ListenAndServe(":"+appPort, mux); err != nil {
		log.Fatalf("Ошибка при запуске HTTP-сервера: %v", err)
	}
