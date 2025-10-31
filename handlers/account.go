package handlers

import (
	"models/database"
	"net/http"

	"github.com/gin-gonic/gin"
)
func GetAccountInfo(c *gin.Context) {
	userID := int(c.MustGet("user_id").(float64))

	var user struct {
		ID       int    `json:"id"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Fullname string `json:"fullname"`
		Phone    string `json:"phone_number"`

	}
	err := database.DB.QueryRow(`SELECT userid, email, username, fullname, phone_number FROM users WHERE userid = $1`, userID).
		Scan(&user.ID, &user.Email, &user.Username, &user.Fullname, &user.Phone)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
