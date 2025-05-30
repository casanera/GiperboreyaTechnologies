// File: internal/models/user.go
package models

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}
