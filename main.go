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
	router.GET("api/hostesses/approved", handlers.GetApprovedHostesses)
	router.GET("api/models/approved", handlers.GetApprovedModels)


			 	
	// ===== USER PROTECTED ROUTES =====
    protected := router.Group("/api", middlewares.AuthMiddleware())
{
    protected.GET("/account", handlers.GetAccountInfo)

    // For Models (User operations)
    protected.POST("/models/create", handlers.CreateModel)
    protected.POST("/models/measurements", handlers.AddMeasurements)
    protected.POST("/models/documents", handlers.AddDocuments)
    protected.POST("/models/identity-check", handlers.UploadIdentityCheck)
    protected.GET("/models/progress", handlers.GetModelProgress)
    protected.DELETE("/models/:id", handlers.DeleteModel)  // User can only delete their own
    protected.PUT("/models/:id", handlers.UpdateModel)     // User can only update their own

    // For Hostesses (User operations)
    protected.POST("/hostesses/create", handlers.CreateHostess)
    protected.POST("/hostesses/experience", handlers.AddHostessExperience)
    protected.POST("/hostesses/documents", handlers.AddHostessDocuments)
    protected.POST("/hostesses/identity-check", handlers.UploadHostessIdentityCheck)
    protected.GET("/hostesses/progress", handlers.GetHostessProgress)
    protected.DELETE("/hostesses/:id", handlers.DeleteHostess)  // User can only delete their own
    protected.PUT("/hostesses/:id", handlers.UpdateHostess)     // User can only update their own
}


// ===== ADMIN PUBLIC ROUTES =====
router.POST("/api/admin/register", handlers.AdminRegister)
router.POST("/api/admin/login", handlers.AdminLogin)


// ===== ADMIN PROTECTED ROUTES =====
adminProtected := router.Group("/api/admin", handlers.AdminAuthMiddleware())
{
    adminProtected.GET("/profile", handlers.GetAdminProfile)
    
    // Admin Model Management
    adminProtected.GET("/models", handlers.AdminGetAllModels)           // Get all models (admin view)
    adminProtected.GET("/models/:id", handlers.AdminGetModelById)       // Get specific model
    adminProtected.POST("/models/:id/approve", handlers.AdminApproveModel)
	adminProtected.PUT("/models/:id", handlers.AdminUpdateModel)  
	adminProtected.POST("/models/:id/reject", handlers.AdminRejectModel)
    adminProtected.DELETE("/models/:id", handlers.AdminDeleteModel)     // Admin can delete any model
    
    // Admin Hostess Management  
    adminProtected.GET("/hostesses", handlers.AdminGetAllHostesses)     // Get all hostesses (admin view)
    adminProtected.GET("/hostesses/:id", handlers.AdminGetHostessById)  // Get specific hostess
	 adminProtected.PUT("/hostesses/:id", handlers.AdminUpdateHostess) 
    adminProtected.DELETE("/hostesses/:id", handlers.AdminDeleteHostess) // Admin can delete any hostess
	adminProtected.POST("/hostesses/:id/approve", handlers.AdminApproveHostess)
	adminProtected.POST("/hostesses/:id/reject", handlers.AdminRejectHostess)
}

// ===== STATIC ROUTES =====
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
