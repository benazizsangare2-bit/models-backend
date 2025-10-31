package handlers

import (
	"fmt"
	"models/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminDeleteModel - Admin can delete any model
func AdminDeleteModel(c *gin.Context) {
    modelID := c.Param("id")

    if modelID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Model ID is required"})
        return
    }

    // Verify the model exists
    var exists bool
    err := database.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM models WHERE id = $1 AND deleted = FALSE)
    `, modelID).Scan(&exists)

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    if !exists {
        c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
        return
    }

    // Soft delete the model
    _, err = database.DB.Exec(`
        UPDATE models SET deleted = TRUE, updated_at = NOW() 
        WHERE id = $1
    `, modelID)

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete model"})
        fmt.Println("Delete error:", err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Model deleted successfully",
    })
}

// AdminDeleteHostess - Admin can delete any hostess
func AdminDeleteHostess(c *gin.Context) {
    hostessID := c.Param("id")

    if hostessID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Hostess ID is required"})
        return
    }

    // Verify the hostess exists
    var exists bool
    err := database.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM hostesses WHERE id = $1 AND deleted = FALSE)
    `, hostessID).Scan(&exists)

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    if !exists {
        c.JSON(http.StatusNotFound, gin.H{"error": "Hostess not found"})
        return
    }

    // Soft delete the hostess
    _, err = database.DB.Exec(`
        UPDATE hostesses SET deleted = TRUE, updated_at = NOW() 
        WHERE id = $1
    `, hostessID)

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete hostess"})
        fmt.Println("Delete error:", err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Hostess deleted successfully",
    })
}