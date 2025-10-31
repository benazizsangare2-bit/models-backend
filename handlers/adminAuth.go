package handlers

import (
	"database/sql"
	"fmt"
	"models/database"

	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)



var adminJWTSecret = []byte(os.Getenv("ADMIN_JWT_SECRET"))

func init() {
    if len(adminJWTSecret) == 0 {
        adminJWTSecret = []byte("admin-secret-key-change-in-production")
    }
}

// AdminRegister handles admin registration
func AdminRegister(c *gin.Context) {
    var req AdminRegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // Check if admin already exists
    var existingAdmin Admin
    err := database.DB.QueryRow(`
        SELECT id FROM admins WHERE (username = $1 OR email = $2) AND deleted = FALSE
    `, req.Username, req.Email).Scan(&existingAdmin.ID)

    if err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Admin with this username or email already exists"})
        return
    } else if err != sql.ErrNoRows {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    // Create admin
    var adminID string
    err = database.DB.QueryRow(`
        INSERT INTO admins (username, email, password_hash, full_name, role)
        VALUES ($1, $2, $3, $4, 'admin')
        RETURNING id
    `, req.Username, req.Email, string(hashedPassword), req.FullName).Scan(&adminID)

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "Admin created successfully",
        "admin_id": adminID,
    })
}

// AdminLogin handles admin authentication
func AdminLogin(c *gin.Context) {
    var req AdminLoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Find admin
    var admin Admin
    err := database.DB.QueryRow(`
        SELECT id, username, email, password_hash, full_name, role, is_active
        FROM admins 
        WHERE username = $1 AND deleted = FALSE
    `, req.Username).Scan(
        &admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash,
        &admin.FullName, &admin.Role, &admin.IsActive,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        }
        return
    }

    // Check if admin is active
    if !admin.IsActive {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin account is deactivated"})
        return
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Update last login
    database.DB.Exec(`UPDATE admins SET last_login = NOW() WHERE id = $1`, admin.ID)

    // Generate JWT token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "admin_id": admin.ID,
        "username": admin.Username,
        "role":     admin.Role,
        "exp":      time.Now().Add(24 * time.Hour).Unix(),
    })

    tokenString, err := token.SignedString(adminJWTSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Login successful",
        "token":   tokenString,
        "admin": gin.H{
            "id":       admin.ID,
            "username": admin.Username,
            "email":    admin.Email,
            "full_name": admin.FullName,
            "role":     admin.Role,
        },
    })
}

// AdminAuthMiddleware for protecting admin routes
func AdminAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
            c.Abort()
            return
        }

        tokenString := authHeader[7:]
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
            }
            return adminJWTSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            c.Abort()
            return
        }

        // Set admin info in context
        c.Set("admin_id", claims["admin_id"])
        c.Set("username", claims["username"])
        c.Set("role", claims["role"])

        c.Next()
    }
}

// GetAdminProfile returns current admin profile
func GetAdminProfile(c *gin.Context) {
    adminID := c.MustGet("admin_id").(string)

    var admin Admin
    err := database.DB.QueryRow(`
        SELECT id, username, email, full_name, role, is_active, last_login, created_at
        FROM admins 
        WHERE id = $1 AND deleted = FALSE
    `, adminID).Scan(
        &admin.ID, &admin.Username, &admin.Email, &admin.FullName,
        &admin.Role, &admin.IsActive, &admin.LastLogin, &admin.CreatedAt,
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch admin profile"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "admin": admin,
    })
}