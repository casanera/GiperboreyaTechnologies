package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/DogeNyan/GiperboreyaTechnologies/internal/models"
	"github.com/DogeNyan/GiperboreyaTechnologies/internal/storage"
)

type UserHandler struct {
	Storage storage.UserStorage
}

func NewUserHandler(s storage.UserStorage) *UserHandler {
	return &UserHandler{Storage: s}
}

func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Ошибка декодирования JSON при создании: %v", err)
		http.Error(w, "Некорректное тело запроса: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if user.Name == "" || user.Email == "" { // Простая валидация
		http.Error(w, "Имя и email обязательны", http.StatusBadRequest)
		return
	}

	id, err := h.Storage.CreateUser(&user)
	if err != nil {
		log.Printf("Ошибка создания пользователя в хранилище: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	user.ID = id

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}
	// Удаляем базовый префикс API, чтобы получить ID или пустую строку
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	idStr = strings.Trim(idStr, "/") // Убираем возможные слеши по краям

	if idStr != "" { // Запрос на конкретного пользователя
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Некорректный ID пользователя '%s': %v", idStr, err)
			http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
			return
		}
		user, err := h.Storage.GetUserByID(id)
		if err != nil {
			if strings.Contains(err.Error(), "не найден") { // Проверяем текст ошибки от storage
				http.Error(w, "Пользователь не найден", http.StatusNotFound)
			} else {
				log.Printf("Ошибка получения пользователя по ID %d: %v", id, err)
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	} else { // Запрос на всех пользователей
		users, err := h.Storage.GetAllUsers()
		if err != nil {
			log.Printf("Ошибка получения всех пользователей: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func (h *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	idStr = strings.Trim(idStr, "/")
	if idStr == "" {
		http.Error(w, "ID пользователя должен быть указан для обновления", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Некорректное тело запроса: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	user.ID = id // Устанавливаем ID из пути

	if user.Name == "" || user.Email == "" {
		http.Error(w, "Имя и email обязательны при обновлении", http.StatusBadRequest)
		return
	}

	err = h.Storage.UpdateUser(&user)
	if err != nil {
		if strings.Contains(err.Error(), "не найден для обновления") {
			http.Error(w, "Пользователь не найден для обновления", http.StatusNotFound)
		} else {
			log.Printf("Ошибка обновления пользователя ID %d: %v", id, err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	idStr = strings.Trim(idStr, "/")
	if idStr == "" {
		http.Error(w, "ID пользователя должен быть указан для удаления", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
		return
	}

	err = h.Storage.DeleteUser(id)
	if err != nil {
		if strings.Contains(err.Error(), "не найден для удаления") {
			http.Error(w, "Пользователь не найден для удаления", http.StatusNotFound)
		} else {
			log.Printf("Ошибка удаления пользователя ID %d: %v", id, err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent) // Успешное удаление
}
