package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/casanera/GiperboreyaTechnologies/internal/models"
	"github.com/casanera/GiperboreyaTechnologies/internal/storage"
)

// setupTest инициализирует UserHandler с MockUserStorage
func setupTest() (*UserHandler, *storage.MockUserStorage) {
	mockStorage := storage.NewMockUserStorage()
	userHandler := NewUserHandler(mockStorage)
	return userHandler, mockStorage
}

func TestCreateUserHandler(t *testing.T) {
	userHandler, mockStorage := setupTest()

	testCases := []struct {
		name               string
		inputPayload       string
		expectedStatusCode int
		expectedName       string // Для проверки данных в ответе
		setupMock          func(*storage.MockUserStorage)
		checkResponse      func(*testing.T, *httptest.ResponseRecorder, string) // Для более детальной проверки тела ответа
	}{
		{
			name:               "Успешное создание",
			inputPayload:       `{"name": "Test User", "email": "test@example.com"}`,
			expectedStatusCode: http.StatusCreated,
			expectedName:       "Test User",
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder, expectedName string) {
				var createdUser models.User
				if err := json.Unmarshal(rr.Body.Bytes(), &createdUser); err != nil {
					t.Fatalf("Не удалось декодировать ответ JSON: %v", err)
				}
				if createdUser.Name != expectedName {
					t.Errorf("Имя пользователя в ответе: ожидалось '%s', получено '%s'", expectedName, createdUser.Name)
				}
				if createdUser.ID == 0 {
					t.Error("ID созданного пользователя не должен быть 0")
				}
			},
		},
		{
			name:               "Некорректный JSON",
			inputPayload:       `{"name": "Bad JSON", "email": "bad@example.com"`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Пустое имя",
			inputPayload:       `{"name": "", "email": "no-name@example.com"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Пустой email",
			inputPayload:       `{"name": "No Email User", "email": ""}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:         "Ошибка хранилища при создании",
			inputPayload: `{"name": "Storage Error", "email": "storage.error@example.com"}`,
			setupMock: func(ms *storage.MockUserStorage) {
				ms.SimulateError = fmt.Errorf("симулированная ошибка БД")
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage.Reset() // Очищаем мок перед каждым тестом
			if tc.setupMock != nil {
				tc.setupMock(mockStorage)
			}

			req, err := http.NewRequest(http.MethodPost, "/api/v1/users/", bytes.NewBufferString(tc.inputPayload))
			if err != nil {
				t.Fatalf("Не удалось создать запрос: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			http.HandlerFunc(userHandler.CreateUserHandler).ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatusCode {
				t.Errorf("Обработчик вернул неверный статус-код: получено %v, ожидалось %v. Тело ответа: %s",
					status, tc.expectedStatusCode, rr.Body.String())
			}

			if tc.checkResponse != nil {
				tc.checkResponse(t, rr, tc.expectedName)
			}
		})
	}
}

func TestGetUserHandler(t *testing.T) {
	userHandler, mockStorage := setupTest()

	// Подготовим данные в моке
	seededUser1 := mockStorage.SeedUser(models.User{Name: "Alice", Email: "alice@example.com"})
	mockStorage.SeedUser(models.User{Name: "Bob", Email: "bob@example.com"})

	t.Run("Получение всех пользователей", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/", nil)
		rr := httptest.NewRecorder()
		http.HandlerFunc(userHandler.GetUserHandler).ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Fatalf("GetAll: неверный статус-код: получено %v, ожидалось %v. Тело: %s", status, http.StatusOK, rr.Body.String())
		}
		var users []models.User
		if err := json.Unmarshal(rr.Body.Bytes(), &users); err != nil {
			t.Fatalf("GetAll: не удалось декодировать JSON: %v", err)
		}
		if len(users) != 2 {
			t.Errorf("GetAll: ожидалось 2 пользователя, получено %d", len(users))
		}
	})

	t.Run("Получение пользователя по ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/"+strconv.FormatInt(seededUser1.ID, 10), nil)
		rr := httptest.NewRecorder()
		http.HandlerFunc(userHandler.GetUserHandler).ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Fatalf("GetByID: неверный статус-код: получено %v, ожидалось %v. Тело: %s", status, http.StatusOK, rr.Body.String())
		}
		var user models.User
		if err := json.Unmarshal(rr.Body.Bytes(), &user); err != nil {
			t.Fatalf("GetByID: не удалось декодировать JSON: %v", err)
		}
		if user.Name != seededUser1.Name {
			t.Errorf("GetByID: имя пользователя неверное, ожидалось '%s', получено '%s'", seededUser1.Name, user.Name)
		}
	})

	t.Run("Получение пользователя по несуществующему ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/999", nil)
		rr := httptest.NewRecorder()
		http.HandlerFunc(userHandler.GetUserHandler).ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("GetByID (not found): неверный статус-код: получено %v, ожидалось %v. Тело: %s", status, http.StatusNotFound, rr.Body.String())
		}
	})

	t.Run("Получение пользователя с некорректным ID (не число)", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/abc", nil)
		rr := httptest.NewRecorder()
		http.HandlerFunc(userHandler.GetUserHandler).ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("GetByID (invalid id format): неверный статус-код: получено %v, ожидалось %v. Тело: %s", status, http.StatusBadRequest, rr.Body.String())
		}
	})
}

func TestUpdateUserHandler(t *testing.T) {
	userHandler, mockStorage := setupTest()
	seededUser := mockStorage.SeedUser(models.User{Name: "Old Name", Email: "old@example.com"})

	updatePayload := `{"name": "New Name", "email": "new@example.com"}`

	t.Run("Успешное обновление", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/"+strconv.FormatInt(seededUser.ID, 10), bytes.NewBufferString(updatePayload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		http.HandlerFunc(userHandler.UpdateUserHandler).ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Fatalf("Update: неверный статус-код: получено %v, ожидалось %v. Тело: %s", status, http.StatusOK, rr.Body.String())
		}
		var updatedUser models.User
		if err := json.Unmarshal(rr.Body.Bytes(), &updatedUser); err != nil {
			t.Fatalf("Update: не удалось декодировать JSON: %v", err)
		}
		if updatedUser.Name != "New Name" || updatedUser.Email != "new@example.com" {
			t.Errorf("Update: данные пользователя не обновились корректно. Получено: %+v", updatedUser)
		}
	})

	t.Run("Обновление несуществующего пользователя", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/999", bytes.NewBufferString(updatePayload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		http.HandlerFunc(userHandler.UpdateUserHandler).ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("Update (not found): неверный статус-код: получено %v, ожидалось %v. Тело: %s", status, http.StatusNotFound, rr.Body.String())
		}
	})
}

func TestDeleteUserHandler(t *testing.T) {
	userHandler, mockStorage := setupTest()
	seededUser := mockStorage.SeedUser(models.User{Name: "To Delete", Email: "delete@example.com"})

	t.Run("Успешное удаление", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/"+strconv.FormatInt(seededUser.ID, 10), nil)
		rr := httptest.NewRecorder()
		http.HandlerFunc(userHandler.DeleteUserHandler).ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("Delete: неверный статус-код: получено %v, ожидалось %v. Тело: %s", status, http.StatusNoContent, rr.Body.String())
		}
		// Проверим, что пользователь действительно удален из мока
		_, err := mockStorage.GetUserByID(seededUser.ID)
		if err == nil || !strings.Contains(err.Error(), "не найден") {
			t.Errorf("Delete: пользователь не был удален из хранилища")
		}
	})

	t.Run("Удаление несуществующего пользователя", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/999", nil)
		rr := httptest.NewRecorder()
		http.HandlerFunc(userHandler.DeleteUserHandler).ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusNotFound { // или StatusInternalServerError, если мок возвращает общую ошибку
			t.Errorf("Delete (not found): неверный статус-код: получено %v, ожидалось %v. Тело: %s", status, http.StatusNotFound, rr.Body.String())
		}
	})
}
