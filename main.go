package main

import (
	"log"
	"models/database"
	"models/handlers"
	middlewares "models/middleware"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
	router := gin.Default() 	
	database.ConnectDatabase()

	// Add CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	    MaxAge: 12 * time.Hour,
	}))
/////////////////// POST ROUTES FOR REGISTRATION  /////////////////////
	router.POST("/register/start", handlers.StartRegistration)  // Step 1: send OTP
	router.POST("/register/verify", handlers.VerifyEmail)       // Step 2: verify OTP
	router.POST("/register/complete", handlers.CompleteRegistration) // Step 3: complete registration
	router.POST("/login", handlers.LoginFunction)
	router.POST("/contact", handlers.HandleContact)             // Contact form submission
			 	
	// middleware protects the route
      protected := router.Group("/api", middlewares.AuthMiddleware())
{
      protected.GET("/account", handlers.GetAccountInfo)
	  protected.POST("/models/create", handlers.CreateModel)
	  protected.POST("/models/measurements", handlers.AddMeasurements)
	  protected.POST("/models/documents", handlers.AddDocuments)
	  protected.POST("/models/identity-check", handlers.UploadIdentityCheck)
	  protected.GET("/models/progress", handlers.GetModelProgress)

	  protected.GET("/models/approved", handlers.GetApprovedModels)
	  protected.DELETE("/models/:id", handlers.DeleteModel)
	  protected.PUT("/models/:id", handlers.UpdateModel)
	  protected.GET("/admin/models", handlers.AdminGetAllModels)
	  protected.GET("/admin/models/:id", handlers.AdminGetModelById)
	  protected.POST("/admin/models/:id/:action", handlers.AdminApproveRejectModel)

	  // For Hostesses
	  protected.POST("/hostesses/create", handlers.CreateHostess)
	  protected.POST("/hostesses/experience", handlers.AddHostessExperience)
	  protected.POST("/hostesses/documents", handlers.AddHostessDocuments)
	  protected.POST("/hostesses/identity-check", handlers.UploadHostessIdentityCheck)
	  protected.GET("/hostesses/progress", handlers.GetHostessProgress)
	  protected.DELETE("/hostesses/:id", handlers.DeleteHostess)
	  protected.PUT("/hostesses/:id", handlers.UpdateHostess)
	  protected.GET("/admin/hostesses", handlers.AdminGetAllHostesses)
	  protected.GET("/admin/hostesses/:id", handlers.AdminGetHostessById)
	  protected.POST("/admin/hostesses/:id/:action", handlers.AdminApproveRejectHostess)


}
  /////////////////// STATIC ROUTES /////////////////////
	router.Static("/uploads", "./uploads") // so uploaded files can be accessed
	
	db := &database.Database{DB: database.DB}
	db.InitDatabase()
	router.GET("/api/", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "Welcome to the api",
		})
	})
	err = router.Run(":6061")
	if err != nil {
		panic(err)
	}
	
}