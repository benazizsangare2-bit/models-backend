package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"fmt"
	"math/rand"
	"models/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"os"
	"regexp"
	"strings"

	"github.com/mailjet/mailjet-apiv3-go"
	"golang.org/x/crypto/bcrypt"
)

type User  struct {
    UserID                int        `json:"userid"`
    Email             	  string     `json:"email"`
    EmailVerified     	  bool       `json:"email_verified"`
    VerificationCode  	  string     `json:"-"`
    VerificationExpiry 	  *time.Time `json:"-"`
    Password          	  string     `json:"password,omitempty"`
    Fullname          	  string     `json:"fullname"`
    Username         	  string     `json:"username"`
    PhoneNumber      	  string     `json:"phone_number"`
    Position         	  string     `json:"position"`
    CreatedAt        	  time.Time  `json:"created_at"`
}


// Function to validate phone number
func isValidPhoneNumber(phonenumber string) bool {
	// Check if the phone number has exactly 10 digits and starts with "07"
	if len(phonenumber) != 10 || !strings.HasPrefix(phonenumber, "07") {
		return false
	}
	return true
}


// Function to validate email format using regex
func isValidEmail(email string) bool {
	const emailRegexPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(emailRegexPattern)
	return regex.MatchString(email)
}

func StartRegistration(ctx *gin.Context) {
    var req struct {
        Email string `json:"email"`
    }

    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    // Normalize email
    req.Email = strings.ToLower(req.Email)

    // Validate email format
    if !isValidEmail(req.Email) {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
        return
    }

    // Check if email already exists
    var exists bool
    err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if exists {
        ctx.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
        return
    }

    // Generate OTP
    otp := fmt.Sprintf("%06d", rand.Intn(1000000))
    expiry := time.Now().Add(10 * time.Minute)

    // Insert new row with OTP (email only for now)
    _, err = database.DB.Exec(
        `INSERT INTO users (email, verification_code, verification_expiry) 
         VALUES ($1, $2, $3) 
         ON CONFLICT (email) DO UPDATE SET verification_code=$2, verification_expiry=$3`,
        req.Email, otp, expiry,
    )
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OTP"})
        return
    }

    // Send OTP to email
    err = sendVerificationEmail(req.Email, otp)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"message": "Verification code sent"})
}

func sendVerificationEmail(Email, otp string) error {
	// NEW: Mailjet implementation
	apiKey := os.Getenv("MAILJET_API_KEY")
	apiSecret := os.Getenv("MAILJET_API_SECRET")
	senderEmail := os.Getenv("MAILJET_SENDER_EMAIL")
	senderName := os.Getenv("MAILJET_SENDER_NAME")

	if apiKey == "" || apiSecret == "" || senderEmail == "" {
		return fmt.Errorf("Mailjet credentials not configured")
	}

	mailjetClient := mailjet.NewMailjetClient(apiKey, apiSecret)

	subject := "Password Reset Request"
	body := fmt.Sprintf("You requested a password reset. Use the code below to reset your password:\n\n%s", otp)

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: senderEmail,
				Name:  senderName,
			},
			To: &mailjet.RecipientsV31{
				{
					Email: Email,
				},
			},
			Subject:  subject,
			TextPart: body,
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := mailjetClient.SendMailV31(&messages)
	if err != nil {
		return fmt.Errorf("Unable to send email via Mailjet: %v", err)
	}

	fmt.Println("Email sent successfully via Mailjet!")
	return nil
}

///////////////// VERIFY OTP ///////////////
func VerifyEmail(ctx *gin.Context) {
    var req struct {
        Email string `json:"email"`
        Code  string `json:"code"`
    }

    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    var dbCode string
    var expiry time.Time
    err := database.DB.QueryRow(
        "SELECT verification_code, verification_expiry FROM users WHERE email = $1",
        req.Email,
    ).Scan(&dbCode, &expiry)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Email not found"})
        return
    }

    if time.Now().After(expiry) {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Verification code expired"})
        return
    }

    if req.Code != dbCode {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid code"})
        return
    }

    // Mark email as verified
    _, err = database.DB.Exec("UPDATE users SET email_verified = TRUE WHERE email = $1", req.Email)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}


/////////// COMPLETE REGISTRATION /////////////////
func CompleteRegistration(ctx *gin.Context) {
    var req struct {
        Email       string `json:"email"`
        Fullname    string `json:"fullname"`
        Username    string `json:"username"`
        PhoneNumber string `json:"phone_number"`
        Password    string `json:"password"`
        Position    string `json:"position"`
    }

    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    // Validate phone number (10 digits, starts with 07)
    if !isValidPhoneNumber(req.PhoneNumber) {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number"})
        return
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    // Update the verified user with details
    _, err = database.DB.Exec(
        `UPDATE users SET fullname=$1, username=$2, phone_number=$3, password=$4, position=$5
         WHERE email=$6 AND email_verified=TRUE`,
        req.Fullname, req.Username, req.PhoneNumber, string(hashedPassword), req.Position, req.Email,
    )
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete registration"})
        fmt.Println("failed to complete registration:", err)
        return
    }

		// Send welcome email to user
		err = sendWelcomEmail(req.Email)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending welcome email"})
			return
		}

    ctx.JSON(http.StatusOK, gin.H{"message": "Registration completed successfully"})
}


/////////// Mailjet to send the welcome email ////////////////////
func sendWelcomEmail(userEmail string) error {
	apiKey := os.Getenv("MAILJET_API_KEY")
	apiSecret := os.Getenv("MAILJET_API_SECRET")
	senderEmail := os.Getenv("MAILJET_SENDER_EMAIL")
	senderName := os.Getenv("MAILJET_SENDER_NAME")

	if apiKey == "" || apiSecret == "" || senderEmail == "" {
		return fmt.Errorf("mailjet credentials not configured")
	}

	mailjetClient := mailjet.NewMailjetClient(apiKey, apiSecret)

	subject := "WELCOME TO YOU!"
	body := "Welcome to Models & Hostesses! We're excited to have you join our community of figures. We look forward to serving you and providing you with an exceptional Service for your events."

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: senderEmail,
				Name:  senderName,
			},
			To: &mailjet.RecipientsV31{
				{
					Email: userEmail,
				},
			},
			Subject:  subject,
			TextPart: body,
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := mailjetClient.SendMailV31(&messages)
	if err != nil {
		return fmt.Errorf("Unable to send email via Mailjet: %v", err)
	}

	fmt.Println("Email sent successfully via Mailjet!")
	return nil
}



//////// Login function with jwt and middleware for protection //////////////

//var jwtSecret = []byte("your_secret_key") // 🔒 Change this later!
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))


// Struct for incoming JSON
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handler
func LoginFunction(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var user User

	// ✅ Query your user table
	err := database.DB.QueryRow(`
		SELECT userid, email, fullname, username, password, created_at
		FROM users 
		WHERE email = $1
	`, req.Email).Scan(&user.UserID, &user.Email, &user.Fullname, &user.Username, &user.Password, &user.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// ✅ Compare password hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// ✅ Generate JWT token (24-hour expiration)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.UserID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// ✅ Return token and user info
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":       user.UserID,
			"email":    user.Email,
			"fullname": user.Fullname,
			"username": user.Username,
			"created":  user.CreatedAt,
		},
	})
}
