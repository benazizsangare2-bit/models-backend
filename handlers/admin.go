package handlers

import (
	"time"
)

type Admin struct {
    ID           string    `json:"id"`
    Username     string    `json:"username"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"`
    FullName     string    `json:"full_name"`
    Role         string    `json:"role"`
    IsActive     bool      `json:"is_active"`
    LastLogin    time.Time `json:"last_login,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    Deleted      bool      `json:"deleted"`
}

type AdminLoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type AdminRegisterRequest struct {
    Username string `json:"username" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    FullName string `json:"full_name" binding:"required"`
}