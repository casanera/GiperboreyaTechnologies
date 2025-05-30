package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/casanera/GiperboreyaTechnologies/internal/models"
	"github.com/casanera/GiperboreyaTechnologies/internal/storage"
)

type UserHandler struct {
	Storage storage.UserStorage
}

func NewUserHandler(s storage.UserStorage) *UserHandler {
	return &UserHandler{Storage: s}
}

// sendJSONResponse вспомогательная функция для отправки JSON ответа
func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil { // Отправляем тело, только если data не nil
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("!!! ОШИБКА кодирования JSON ответа: %v. Данные: %+v", err, data)
		} else {
			log.Printf("DEBUG: JSON ответ успешно отправлен. Статус: %d. Данные: %+v", statusCode, data)
		}
	} else {
		// Если data is nil, но статус не 204, это может быть намеренно (например, ошибка обработана ранее)
		// или это может быть 204 No Content, где тело и не нужно.
		// Для 204 WriteHeader(http.StatusNoContent) достаточно.
		log.Printf("DEBUG: JSON ответ отправлен. Статус: %d. Тело ответа: nil (или не предполагалось)", statusCode)
	}
}

// sendErrorResponse вспомогательная функция для отправки JSON ошибки
func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	log.Printf("Отправка ошибки: Статус %d, Сообщение: %s", statusCode, message)
	w.Header().Set("Content-Type", "application/json") // Убедимся, что даже ошибки в JSON
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DEBUG: CreateUserHandler - Начало обработки")
	if r.Method != http.MethodPost {
		sendErrorResponse(w, http.StatusMethodNotAllowed, "Метод не разрешен")
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Ошибка декодирования JSON при создании: %v. Тело запроса: %v", err, r.Body)
		sendErrorResponse(w, http.StatusBadRequest, "Некорректное тело запроса: "+err.Error())
		return
	}
	defer r.Body.Close() // Важно закрывать тело запроса

	log.Printf("DEBUG: CreateUserHandler - Декодированные данные пользователя: %+v", user)

	if user.Name == "" || user.Email == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Имя и email обязательны")
		return
	}

	id, err := h.Storage.CreateUser(&user)
	if err != nil {
		log.Printf("Ошибка h.Storage.CreateUser: %v. Пользователь: %+v", err, user)
		sendErrorResponse(w, http.StatusInternalServerError, "Внутренняя ошибка сервера при создании пользователя")
		return
	}
	user.ID = id // Присваиваем ID, полученный от хранилища
	log.Printf("DEBUG: CreateUserHandler - Пользователь создан с ID: %d. Данные: %+v", id, user)

	sendJSONResponse(w, http.StatusCreated, user)
}

func (h *UserHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DEBUG: GetUserHandler - Начало обработки")
	if r.Method != http.MethodGet {
		sendErrorResponse(w, http.StatusMethodNotAllowed, "Метод не разрешен")
		return
	}

	// r.URL.Path здесь будет, например, "/api/v1/users/" или "/api/v1/users/123"
	// `mux.HandleFunc("/api/v1/users/", ...)` уже обеспечил этот префикс
	idStrWithSlashes := strings.TrimPrefix(r.URL.Path, "/api/v1/users") // результат: "/" или "/123" или "/123/"
	idStr := strings.Trim(idStrWithSlashes, "/")                        // результат: "" или "123"

	if idStr != "" { // Запрос на конкретного пользователя
		log.Printf("DEBUG: GetUserHandler - Запрос на пользователя по ID: '%s'", idStr)
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Некорректный ID пользователя '%s': %v", idStr, err)
			sendErrorResponse(w, http.StatusBadRequest, "Некорректный ID пользователя")
			return
		}

		user, err := h.Storage.GetUserByID(id)
		if err != nil {
			if strings.Contains(err.Error(), "не найден") { // Предполагаем, что storage возвращает такую строку
				log.Printf("Пользователь с ID %d не найден в хранилище.", id)
				sendErrorResponse(w, http.StatusNotFound, "Пользователь не найден")
			} else {
				log.Printf("Ошибка h.Storage.GetUserByID для ID %d: %v", id, err)
				sendErrorResponse(w, http.StatusInternalServerError, "Внутренняя ошибка сервера при получении пользователя")
			}
			return
		}
		log.Printf("DEBUG: GetUserHandler - Найден пользователь по ID %d: %+v", id, user)
		sendJSONResponse(w, http.StatusOK, user)

	} else { // Запрос на всех пользователей
		log.Println("DEBUG: GetUserHandler - Запрос на ВСЕХ пользователей")
		users, err := h.Storage.GetAllUsers()
		if err != nil {
			log.Printf("Ошибка h.Storage.GetAllUsers: %v", err)
			sendErrorResponse(w, http.StatusInternalServerError, "Внутренняя ошибка сервера при получении списка пользователей")
			return
		}

		if users == nil { // На всякий случай, хотя storage должен возвращать пустой слайс
			log.Println("DEBUG: GetUserHandler - h.Storage.GetAllUsers() вернул nil, инициализируем пустым слайсом")
			users = []models.User{}
		}
		log.Printf("DEBUG: GetUserHandler - Получено %d пользователей: %+v", len(users), users)
		sendJSONResponse(w, http.StatusOK, users)
	}
}

func (h *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DEBUG: UpdateUserHandler - Начало обработки")
	if r.Method != http.MethodPut {
		sendErrorResponse(w, http.StatusMethodNotAllowed, "Метод не разрешен")
		return
	}

	idStrWithSlashes := strings.TrimPrefix(r.URL.Path, "/api/v1/users")
	idStr := strings.Trim(idStrWithSlashes, "/")
	if idStr == "" {
		sendErrorResponse(w, http.StatusBadRequest, "ID пользователя должен быть указан в пути для обновления")
		return
	}
	log.Printf("DEBUG: UpdateUserHandler - Запрос на обновление пользователя по ID: '%s'", idStr)

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("Некорректный ID пользователя для обновления '%s': %v", idStr, err)
		sendErrorResponse(w, http.StatusBadRequest, "Некорректный ID пользователя")
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Ошибка декодирования JSON при обновлении: %v. Тело запроса: %v", err, r.Body)
		sendErrorResponse(w, http.StatusBadRequest, "Некорректное тело запроса: "+err.Error())
		return
	}
	defer r.Body.Close()
	user.ID = id // Устанавливаем ID из пути, чтобы он был в объекте user
	log.Printf("DEBUG: UpdateUserHandler - Декодированные данные для обновления пользователя ID %d: %+v", id, user)

	if user.Name == "" || user.Email == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Имя и email обязательны при обновлении")
		return
	}

	err = h.Storage.UpdateUser(&user)
	if err != nil {
		if strings.Contains(err.Error(), "не найден для обновления") {
			log.Printf("Пользователь с ID %d не найден для обновления в хранилище.", id)
			sendErrorResponse(w, http.StatusNotFound, "Пользователь не найден для обновления")
		} else {
			log.Printf("Ошибка h.Storage.UpdateUser для ID %d: %v. Данные: %+v", id, err, user)
			sendErrorResponse(w, http.StatusInternalServerError, "Внутренняя ошибка сервера при обновлении пользователя")
		}
		return
	}
	log.Printf("DEBUG: UpdateUserHandler - Пользователь ID %d успешно обновлен. Новые данные: %+v", id, user)
	sendJSONResponse(w, http.StatusOK, user) // Возвращаем обновленного пользователя
}

func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DEBUG: DeleteUserHandler - Начало обработки")
	if r.Method != http.MethodDelete {
		sendErrorResponse(w, http.StatusMethodNotAllowed, "Метод не разрешен")
		return
	}

	idStrWithSlashes := strings.TrimPrefix(r.URL.Path, "/api/v1/users")
	idStr := strings.Trim(idStrWithSlashes, "/")
	if idStr == "" {
		sendErrorResponse(w, http.StatusBadRequest, "ID пользователя должен быть указан в пути для удаления")
		return
	}
	log.Printf("DEBUG: DeleteUserHandler - Запрос на удаление пользователя по ID: '%s'", idStr)

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("Некорректный ID пользователя для удаления '%s': %v", idStr, err)
		sendErrorResponse(w, http.StatusBadRequest, "Некорректный ID пользователя")
		return
	}

	err = h.Storage.DeleteUser(id)
	if err != nil {
		if strings.Contains(err.Error(), "не найден для удаления") {
			log.Printf("Пользователь с ID %d не найден для удаления в хранилище.", id)
			sendErrorResponse(w, http.StatusNotFound, "Пользователь не найден для удаления")
		} else {
			log.Printf("Ошибка h.Storage.DeleteUser для ID %d: %v", id, err)
			sendErrorResponse(w, http.StatusInternalServerError, "Внутренняя ошибка сервера при удалении пользователя")
		}
		return
	}
	log.Printf("DEBUG: DeleteUserHandler - Пользователь ID %d успешно удален.", id)
	w.WriteHeader(http.StatusNoContent)

}
