package storage

import (
	"fmt"
	"sync"

	"github.com/casanera/GiperboreyaTechnologies/internal/models"
)

// MockUserStorage является мок-реализацией UserStorage для тестов
type MockUserStorage struct {
	mu            sync.Mutex
	Users         map[int64]*models.User
	NextID        int64
	SimulateError error
}

// NewMockUserStorage создает новый экземпляр MockUserStorage.
func NewMockUserStorage() *MockUserStorage {
	return &MockUserStorage{
		Users:  make(map[int64]*models.User),
		NextID: 1,
	}
}

func (m *MockUserStorage) CreateUser(user *models.User) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SimulateError != nil {
		return 0, m.SimulateError
	}
	for _, existingUser := range m.Users {
		if existingUser.Email == user.Email {
			return 0, fmt.Errorf("мок: email '%s' уже существует", user.Email)
		}
	}

	newID := m.NextID
	m.NextID++
	user.ID = newID // Присваиваем ID мок-объекту
	userCopy := *user
	m.Users[newID] = &userCopy
	return newID, nil
}

func (m *MockUserStorage) GetUserByID(id int64) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SimulateError != nil {
		return nil, m.SimulateError
	}
	user, exists := m.Users[id]
	if !exists {
		return nil, fmt.Errorf("storage.GetUserByID: пользователь с ID %d не найден", id) // Совпадает с ошибкой в PostgresUserStorage
	}
	userCopy := *user // Возвращаем копию
	return &userCopy, nil
}

func (m *MockUserStorage) GetAllUsers() ([]models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SimulateError != nil {
		return nil, m.SimulateError
	}
	var usersList []models.User
	for _, user := range m.Users {
		userCopy := *user
		usersList = append(usersList, userCopy)
	}
	return usersList, nil
}

func (m *MockUserStorage) UpdateUser(user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SimulateError != nil {
		return m.SimulateError
	}
	// Проверка на существующий email (кроме текущего пользователя)
	for id, existingUser := range m.Users {
		if id != user.ID && existingUser.Email == user.Email {
			return fmt.Errorf("мок: email '%s' уже используется другим пользователем", user.Email)
		}
	}
	_, exists := m.Users[user.ID]
	if !exists {
		return fmt.Errorf("storage.UpdateUser: пользователь с ID %d не найден для обновления", user.ID)
	}
	userCopy := *user
	m.Users[user.ID] = &userCopy
	return nil
}

func (m *MockUserStorage) DeleteUser(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SimulateError != nil {
		return m.SimulateError
	}
	_, exists := m.Users[id]
	if !exists {
		return fmt.Errorf("storage.DeleteUser: пользователь с ID %d не найден для удаления", id)
	}
	delete(m.Users, id)
	return nil
}

// Вспомогательный метод для тестов, чтобы очищать мок между тестами
func (m *MockUserStorage) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Users = make(map[int64]*models.User)
	m.NextID = 1
	m.SimulateError = nil
}

// Вспомогательный метод для добавления пользователя напрямую в мок для настройки тестов
func (m *MockUserStorage) SeedUser(user models.User) models.User {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user.ID == 0 {
		user.ID = m.NextID
		m.NextID++
	} else if user.ID >= m.NextID {
		m.NextID = user.ID + 1
	}
	m.Users[user.ID] = &user
	return user
}
