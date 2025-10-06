package handlers

import (
	"database/sql"
	"fmt"
	"models/database"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// Core model identity
type Model struct {
    ID                  string    `json:"id"`
    UserID              int       `json:"userid"`
    FirstName           string    `json:"first_name"`
    LastName            string    `json:"last_name"`
    Username            string    `json:"username"`
    Email               string    `json:"email"`
    Whatsapp            string    `json:"whatsapp"`
    DateOfBirth         string    `json:"date_of_birth"`
    Gender              string    `json:"gender"`
    Nationality         string    `json:"nationality"`
    Street              string    `json:"street"`
    City                string    `json:"city"`
    ResidenceCountry    string    `json:"residence_country"`
    Status              string    `json:"status"`
    CreatedAt           time.Time `json:"created_at"`
    UpdatedAt           time.Time `json:"updated_at"`
}

// Measurements
type ModelMeasurements struct {
    ID                  string `json:"id"`
    ModelID             string `json:"model_id"`
    Experience          string `json:"experience"`
    Height              int    `json:"height"`
    Weight              int    `json:"weight"`
    Hips                int    `json:"hips"`
    HairColor           string `json:"hair_color"`
    Waist               int    `json:"waist"`
    EyeColor            string `json:"eye_color"`
    Photo               string `json:"photo"` // Store as a comma-separated string
    Additional_photo    string `json:"additional_photo"`
}

// Documents
type ModelDocuments struct {
    ID                    string    `json:"id"`
    ModelID               string    `json:"model_id"`
    DocumentIssuerCountry string    `json:"documentIssuerCountry"`
    DocumentType          string    `json:"documentType"`
    DocumentFront         string    `json:"documentFront"`
    DocumentBack          string    `json:"documentBack"`
    CreatedAt             time.Time `json:"created_at"`
    UpdatedAt             time.Time `json:"updated_at"`
}



// PERSONAL INFO (Step 1 of registration)
func CreateModel(ctx *gin.Context) {
	var req Model

	// Parse request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		fmt.Println("Invalid input:", err)
		return
	}

	// Get logged-in user details from context (set in AuthMiddleware)
	userID := int(ctx.MustGet("user_id").(float64))
	email := ctx.MustGet("email").(string)

	// Force link model to logged-in user
	req.UserID = userID
	req.Email = email

	// Validate required field
	if req.DateOfBirth == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Date of birth is required"})
		return
	}

	// Validate DOB format
	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, expected YYYY-MM-DD"})
		return
	}
	// Insert into database
	query := `
		INSERT INTO models (
			first_name, last_name, username, email, whatsapp, date_of_birth,
			gender, nationality, street, city, residence_country, user_id, created_at, updated_at, deleted
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW(),NOW(),FALSE)
		RETURNING id
	`

	err = database.DB.QueryRow(
		query,
		req.FirstName, req.LastName, req.Username, req.Email, req.Whatsapp,
		dob, req.Gender, req.Nationality, req.Street, req.City, req.ResidenceCountry,req.UserID,
	).Scan(&req.ID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create model"})
		fmt.Println("Failed to create model:", err)
		return
	}

	_, _ = database.DB.Exec(`UPDATE models SET registration_step = 1 WHERE id = $1`, req.ID)

ctx.JSON(http.StatusOK, gin.H{
    "id": req.ID,
    "step": 1,
    "message": "Step 1 saved successfully",
})

}


///////// MEASUREMENTS ////////
func AddMeasurements(ctx *gin.Context) {
    // Parse form fields
    modelID := ctx.PostForm("model_id")
    experience := ctx.PostForm("experience")
    height := ctx.PostForm("height")
    weight := ctx.PostForm("weight")
    hips := ctx.PostForm("hips")
    waist := ctx.PostForm("waist")
    hairColor := ctx.PostForm("hair_color")
    eyeColor := ctx.PostForm("eye_color")

    // Validate required fields
    if modelID == "" || experience == "" || height == "" || weight == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
        return
    }

    // Ensure upload folder exists
    uploadDir := "uploads/measurements"
    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
            return
        }
    }

    // ---- Handle main photo ----
    var mainPhotoPath string
    mainPhoto, err := ctx.FormFile("photo")
    if err == nil {
        mainPhotoPath = fmt.Sprintf("%s/%d_%s", uploadDir, time.Now().UnixNano(), mainPhoto.Filename)
        if err := ctx.SaveUploadedFile(mainPhoto, mainPhotoPath); err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save main photo"})
            return
        }
    }

    // ---- Handle additional photos (max 5) ----
    var additionalPhotoPaths []string
    form, err := ctx.MultipartForm()
    if err == nil && form.File["additionalPhotos"] != nil {
        photos := form.File["additionalPhotos"]
        if len(photos) > 5 {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 5 additional photos allowed"})
            return
        }

        for _, file := range photos {
            path := fmt.Sprintf("%s/%d_%s", uploadDir, time.Now().UnixNano(), file.Filename)
            if err := ctx.SaveUploadedFile(file, path); err != nil {
                ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save additional photo"})
                return
            }
            additionalPhotoPaths = append(additionalPhotoPaths, path)
        }
    }

    // ---- Insert into DB ----
    query := `
        INSERT INTO model_measurements 
        (model_id, experience, height, weight, hips, waist, hair_color, eye_color, photo, additional_photos)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
    `
    _, err = database.DB.Exec(query,
        modelID, experience, height, weight, hips, waist,
        hairColor, eyeColor, mainPhotoPath, pq.Array(additionalPhotoPaths),
    )
    if err != nil {
        fmt.Println("DB insert error:", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save measurements"})
        return
    }

    _, _ = database.DB.Exec(`UPDATE models SET registration_step = 2 WHERE id = $1`, modelID)

    ctx.JSON(http.StatusOK, gin.H{"message": "Step 2 saved successfully"})
}



/// Document Verification 
func AddDocuments(ctx *gin.Context) {

    modelID := ctx.PostForm("model_id")
    documentIssuerCountry := ctx.PostForm("documentIssuerCountry")
    documentType := ctx.PostForm("documentType")

    if modelID == "" || documentIssuerCountry == "" || documentType == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
        return
    }

    frontFile, err := ctx.FormFile("documentFront")
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Front document image is required"})
        return
    }

    backFile, err := ctx.FormFile("documentBack")
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Back document image is required"})
        return
    }


    uploadDir := "uploads"
if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
    if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
        return
    }
}

    frontPath := fmt.Sprintf("uploads/%d_%s", time.Now().Unix(), frontFile.Filename)
    backPath := fmt.Sprintf("uploads/%d_%s", time.Now().Unix(), backFile.Filename)

    if err := ctx.SaveUploadedFile(frontFile, frontPath); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save front image"})
        return
    }

    if err := ctx.SaveUploadedFile(backFile, backPath); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save back image"})
        return
    }

    // Insert document verification entry
    query := `
        INSERT INTO model_documents (model_id, document_issuer_country, document_type, document_front, document_back)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
    var docID string
    err = database.DB.QueryRow(query, modelID, documentIssuerCountry, documentType, frontPath, backPath).Scan(&docID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save documents"})
        fmt.Println("Database error:", err)
        return
    }

    _, _ = database.DB.Exec(`UPDATE models SET registration_step = 3 WHERE id = $1`, modelID)


    ctx.JSON(http.StatusOK, gin.H{
        "id":      docID,
        "message": "Step 3 saved successfully",
    })
}


// Identity Check - Selfie Upload
func UploadIdentityCheck(ctx *gin.Context) {
    modelID := ctx.PostForm("model_id")
    if modelID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model ID is required"})
        return
    }

    file, err := ctx.FormFile("selfie_with_id")
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Selfie with ID file is required"})
        return
    }

    // Ensure uploads folder exists
    uploadDir := "uploads/identity_check"
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
        return
    }

    // Save selfie file
    filename := fmt.Sprintf("%s_%d_%s", modelID, time.Now().Unix(), file.Filename)
    savePath := filepath.Join(uploadDir, filename)
    if err := ctx.SaveUploadedFile(file, savePath); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save selfie with ID"})
        return
    }

    // Insert record into database
    query := `INSERT INTO model_identity_check (model_id, selfie_with_id) VALUES ($1, $2)`
    _, err = database.DB.Exec(query, modelID, savePath)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database insert failed"})
        return
    }
    _, _ = database.DB.Exec(`UPDATE models SET registration_step = 4 WHERE id = $1`, modelID)

    ctx.JSON(http.StatusOK, gin.H{
        "message": "Selfie with ID uploaded successfully",
        "file": savePath,
    })
}

/// User's last progress
func GetModelProgress(ctx *gin.Context) {
    userID := int(ctx.MustGet("user_id").(float64))

    var modelID string
    var step int
    err := database.DB.QueryRow(`
        SELECT id, registration_step
        FROM models
        WHERE user_id = $1 AND deleted = FALSE
        ORDER BY created_at DESC
        LIMIT 1
    `, userID).Scan(&modelID, &step)

    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "No model found for user"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "model_id": modelID,
        "current_step": step,
    })
}

// Get all approved models with full information (for the gallery)
func GetApprovedModels(ctx *gin.Context) {
    query := `
        SELECT 
            m.id, m.user_id, m.first_name, m.last_name, m.username, m.email, 
            m.whatsapp, m.date_of_birth, m.gender, m.nationality, m.street, 
            m.city, m.residence_country, m.status, m.created_at, m.updated_at,
            mm.experience, mm.height, mm.weight, mm.hips, mm.waist, 
            mm.hair_color, mm.eye_color, mm.photo, mm.additional_photos,
            md.document_issuer_country, md.document_type, md.document_front, md.document_back,
            mic.selfie_with_id, mic.verified as identity_verified
        FROM models m
        LEFT JOIN model_measurements mm ON m.id = mm.model_id
        LEFT JOIN model_documents md ON m.id = md.model_id
        LEFT JOIN model_identity_check mic ON m.id = mic.model_id
        WHERE m.status = 'approved' AND m.deleted = FALSE
        ORDER BY m.created_at DESC
    `

    rows, err := database.DB.Query(query)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approved models"})
        fmt.Println("Database query error:", err)
        return
    }
    defer rows.Close()

    var models []gin.H
    for rows.Next() {
        var model gin.H
        var (
            id, userID, firstName, lastName, username, email, whatsapp string
            dateOfBirth, gender, nationality, street, city, residenceCountry, status string
            createdAt, updatedAt time.Time
            experience, hairColor, eyeColor, photo, additionalPhotos sql.NullString
            height, weight, hips, waist sql.NullInt64
            docIssuerCountry, docType, docFront, docBack sql.NullString
            selfieWithID sql.NullString
            identityVerified sql.NullBool
        )

        err := rows.Scan(
            &id, &userID, &firstName, &lastName, &username, &email, &whatsapp,
            &dateOfBirth, &gender, &nationality, &street, &city, &residenceCountry, &status,
            &createdAt, &updatedAt, &experience, &height, &weight, &hips, &waist,
            &hairColor, &eyeColor, &photo, &additionalPhotos, &docIssuerCountry,
            &docType, &docFront, &docBack, &selfieWithID, &identityVerified,
        )
        if err != nil {
            fmt.Println("Row scan error:", err)
            continue
        }

        // Helper function to get string value or default
        getStringValue := func(value sql.NullString, defaultValue string) string {
            if value.Valid && value.String != "" {
                return value.String
            }
            return defaultValue
        }

        // Helper function to get int64 value or default
        getIntValue := func(value sql.NullInt64, defaultValue int64) int64 {
            if value.Valid {
                return value.Int64
            }
            return defaultValue
        }

        // Helper function to get bool value or default
        getBoolValue := func(value sql.NullBool, defaultValue bool) bool {
            if value.Valid {
                return value.Bool
            }
            return defaultValue
        }

        // Get photo path or provide default
        photoPath := getStringValue(photo, "")
        if photoPath == "" {
            photoPath = "/uploads/default.jpg" // Default image path using existing uploads folder
        }

        // Get additional photos as array
        var additionalPhotosArray []string
        if additionalPhotos.Valid && additionalPhotos.String != "" {
            additionalPhotosArray = strings.Split(additionalPhotos.String, ",")
        }

        model = gin.H{
            "id": id,
            "user_id": userID,
            "first_name": firstName,
            "last_name": lastName,
            "username": username,
            "email": email,
            "whatsapp": whatsapp,
            "date_of_birth": dateOfBirth,
            "gender": gender,
            "nationality": nationality,
            "street": street,
            "city": city,
            "residence_country": residenceCountry,
            "status": status,
            "created_at": createdAt,
            "updated_at": updatedAt,
            // Flatten measurements for easier frontend access
            "experience": getStringValue(experience, ""),
            "height": getIntValue(height, 0),
            "weight": getIntValue(weight, 0),
            "hips": getIntValue(hips, 0),
            "waist": getIntValue(waist, 0),
            "hair_color": getStringValue(hairColor, ""),
            "eye_color": getStringValue(eyeColor, ""),
            "photo": photoPath,
            "additional_photos": additionalPhotosArray,
            "measurements": gin.H{
                "experience": getStringValue(experience, ""),
                "height": getIntValue(height, 0),
                "weight": getIntValue(weight, 0),
                "hips": getIntValue(hips, 0),
                "waist": getIntValue(waist, 0),
                "hair_color": getStringValue(hairColor, ""),
                "eye_color": getStringValue(eyeColor, ""),
                "photo": photoPath,
                "additional_photos": additionalPhotosArray,
            },
            "documents": gin.H{
                "document_issuer_country": getStringValue(docIssuerCountry, ""),
                "document_type": getStringValue(docType, ""),
                "document_front": getStringValue(docFront, ""),
                "document_back": getStringValue(docBack, ""),
            },
            "identity_check": gin.H{
                "selfie_with_id": getStringValue(selfieWithID, ""),
                "verified": getBoolValue(identityVerified, false),
            },
        }
        models = append(models, model)
    }

    ctx.JSON(http.StatusOK, gin.H{
        "models": models,
        "count": len(models),
    })
}

// Soft delete model by setting deleted = true
func DeleteModel(ctx *gin.Context) {
    userID := int(ctx.MustGet("user_id").(float64))
    modelID := ctx.Param("id")

    if modelID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model ID is required"})
        return
    }

    // Verify the model belongs to the user
    var ownerID int
    err := database.DB.QueryRow(`
        SELECT user_id FROM models 
        WHERE id = $1 AND deleted = FALSE
    `, modelID).Scan(&ownerID)

    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
        return
    }

    if ownerID != userID {
        ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own model"})
        return
    }

    // Soft delete the model
    _, err = database.DB.Exec(`
        UPDATE models SET deleted = TRUE, updated_at = NOW() 
        WHERE id = $1
    `, modelID)

    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete model"})
        fmt.Println("Delete error:", err)
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "message": "Model deleted successfully",
    })
}

// Update model information
func UpdateModel(ctx *gin.Context) {
    userID := int(ctx.MustGet("user_id").(float64))
    modelID := ctx.Param("id")

    if modelID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model ID is required"})
        return
    }

    // Verify the model belongs to the user
    var ownerID int
    err := database.DB.QueryRow(`
        SELECT user_id FROM models 
        WHERE id = $1 AND deleted = FALSE
    `, modelID).Scan(&ownerID)

    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
        return
    }

    if ownerID != userID {
        ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own model"})
        return
    }

    // Parse request body
    var req Model
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        fmt.Println("Invalid input:", err)
        return
    }

    // Validate DOB format if provided
    var dob interface{}
    if req.DateOfBirth != "" {
        parsedDob, err := time.Parse("2006-01-02", req.DateOfBirth)
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, expected YYYY-MM-DD"})
            return
        }
        dob = parsedDob
    }

    // Update model information
    query := `
        UPDATE models SET 
            first_name = COALESCE($2, first_name),
            last_name = COALESCE($3, last_name),
            username = COALESCE($4, username),
            whatsapp = COALESCE($5, whatsapp),
            date_of_birth = COALESCE($6, date_of_birth),
            gender = COALESCE($7, gender),
            nationality = COALESCE($8, nationality),
            street = COALESCE($9, street),
            city = COALESCE($10, city),
            residence_country = COALESCE($11, residence_country),
            updated_at = NOW()
        WHERE id = $1
    `

    _, err = database.DB.Exec(query,
        modelID,
        req.FirstName, req.LastName, req.Username, req.Whatsapp,
        dob, req.Gender, req.Nationality, req.Street, req.City, req.ResidenceCountry,
    )

    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update model"})
        fmt.Println("Update error:", err)
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "message": "Model updated successfully",
        "model_id": modelID,
    })
}

// Admin endpoint to get all models with complete information for review
func AdminGetAllModels(ctx *gin.Context) {
    // Get query parameters for filtering
    status := ctx.Query("status") // pending, under_review, approved, rejected
    page := ctx.DefaultQuery("page", "1")
    limit := ctx.DefaultQuery("limit", "10")

    // Build base query
    baseQuery := `
        SELECT 
            m.id, m.user_id, m.first_name, m.last_name, m.username, m.email, 
            m.whatsapp, m.date_of_birth, m.gender, m.nationality, m.street, 
            m.city, m.residence_country, m.status, m.registration_step, m.deleted,
            m.created_at, m.updated_at,
            mm.experience, mm.height, mm.weight, mm.hips, mm.waist, 
            mm.hair_color, mm.eye_color, mm.photo, mm.additional_photos,
            md.document_issuer_country, md.document_type, md.document_front, md.document_back,
            mic.selfie_with_id, mic.verified as identity_verified,
            u.fullname as user_fullname, u.email as user_email, u.phone_number as user_phone
        FROM models m
        LEFT JOIN model_measurements mm ON m.id = mm.model_id
        LEFT JOIN model_documents md ON m.id = md.model_id
        LEFT JOIN model_identity_check mic ON m.id = mic.model_id
        LEFT JOIN users u ON m.user_id = u.userid
        WHERE m.deleted = FALSE
    `

    // Add status filter if provided
    if status != "" {
        baseQuery += " AND m.status = $1"
    }
    
    baseQuery += " ORDER BY m.created_at DESC"

    // Execute query
    var rows *sql.Rows
    var err error
    
    if status != "" {
        rows, err = database.DB.Query(baseQuery, status)
    } else {
        rows, err = database.DB.Query(baseQuery)
    }

    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch models"})
        fmt.Println("Database query error:", err)
        return
    }
    defer rows.Close()

    var models []gin.H
    for rows.Next() {
        var (
            id, userID, firstName, lastName, username, email, whatsapp string
            dateOfBirth, gender, nationality, street, city, residenceCountry, status string
            createdAt, updatedAt time.Time
            registrationStep int
            deleted bool
            experience, hairColor, eyeColor, photo, additionalPhotos sql.NullString
            height, weight, hips, waist sql.NullInt64
            docIssuerCountry, docType, docFront, docBack sql.NullString
            selfieWithID sql.NullString
            identityVerified sql.NullBool
            userFullname, userEmail, userPhone sql.NullString
        )

        err := rows.Scan(
            &id, &userID, &firstName, &lastName, &username, &email, &whatsapp,
            &dateOfBirth, &gender, &nationality, &street, &city, &residenceCountry, &status,
            &registrationStep, &deleted, &createdAt, &updatedAt, &experience, &height, &weight, 
            &hips, &waist, &hairColor, &eyeColor, &photo, &additionalPhotos, &docIssuerCountry,
            &docType, &docFront, &docBack, &selfieWithID, &identityVerified,
            &userFullname, &userEmail, &userPhone,
        )
        if err != nil {
            fmt.Println("Row scan error:", err)
            continue
        }

        model := gin.H{
            "id": id,
            "user_id": userID,
            "first_name": firstName,
            "last_name": lastName,
            "username": username,
            "email": email,
            "whatsapp": whatsapp,
            "date_of_birth": dateOfBirth,
            "gender": gender,
            "nationality": nationality,
            "street": street,
            "city": city,
            "residence_country": residenceCountry,
            "status": status,
            "registration_step": registrationStep,
            "deleted": deleted,
            "created_at": createdAt,
            "updated_at": updatedAt,
            "user_info": gin.H{
                "fullname": userFullname.String,
                "email": userEmail.String,
                "phone": userPhone.String,
            },
            "measurements": gin.H{
                "experience": experience.String,
                "height": height.Int64,
                "weight": weight.Int64,
                "hips": hips.Int64,
                "waist": waist.Int64,
                "hair_color": hairColor.String,
                "eye_color": eyeColor.String,
                "photo": photo.String,
                "additional_photos": additionalPhotos.String,
            },
            "documents": gin.H{
                "document_issuer_country": docIssuerCountry.String,
                "document_type": docType.String,
                "document_front": docFront.String,
                "document_back": docBack.String,
            },
            "identity_check": gin.H{
                "selfie_with_id": selfieWithID.String,
                "verified": identityVerified.Bool,
            },
        }
        models = append(models, model)
    }

    // Get total count for pagination
    countQuery := "SELECT COUNT(*) FROM models WHERE deleted = FALSE"
    var totalCount int
    if status != "" {
        err = database.DB.QueryRow(countQuery + " AND status = $1", status).Scan(&totalCount)
    } else {
        err = database.DB.QueryRow(countQuery).Scan(&totalCount)
    }

    if err != nil {
        fmt.Println("Count query error:", err)
        totalCount = len(models)
    }

    ctx.JSON(http.StatusOK, gin.H{
        "models": models,
        "total_count": totalCount,
        "filtered_count": len(models),
        "status_filter": status,
        "page": page,
        "limit": limit,
    })
}

// Admin endpoint to approve or reject a model
func AdminApproveRejectModel(ctx *gin.Context) {
    modelID := ctx.Param("id")
    action := ctx.Param("action") // approve or reject

    if modelID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model ID is required"})
        return
    }

    if action != "approve" && action != "reject" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Action must be 'approve' or 'reject'"})
        return
    }

    // Verify the model exists and is not deleted
    var currentStatus string
    err := database.DB.QueryRow(`
        SELECT status FROM models 
        WHERE id = $1 AND deleted = FALSE
    `, modelID).Scan(&currentStatus)

    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
        return
    }

    // Check if model is already in final state
    if currentStatus == "approved" || currentStatus == "rejected" {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": fmt.Sprintf("Model is already %s", currentStatus),
        })
        return
    }

    // Parse request body for admin notes (optional)
    var req struct {
        AdminNotes string `json:"admin_notes"`
    }
    ctx.ShouldBindJSON(&req)

    // Update model status
    var newStatus string
    if action == "approve" {
        newStatus = "approved"
    } else {
        newStatus = "rejected"
    }

    query := `
        UPDATE models 
        SET status = $1, updated_at = NOW() 
        WHERE id = $2
    `
    
    _, err = database.DB.Exec(query, newStatus, modelID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update model status"})
        fmt.Println("Update error:", err)
        return
    }

    // Log the action (you might want to create a separate admin_logs table for this)
    logMessage := fmt.Sprintf("Model %s %s by admin", modelID, newStatus)
    if req.AdminNotes != "" {
        logMessage += fmt.Sprintf(" - Notes: %s", req.AdminNotes)
    }
    fmt.Println(logMessage)

    ctx.JSON(http.StatusOK, gin.H{
        "message": fmt.Sprintf("Model %s successfully", newStatus),
        "model_id": modelID,
        "new_status": newStatus,
        "admin_notes": req.AdminNotes,
    })
}

// Admin endpoint to get a specific model by ID with complete information
func AdminGetModelById(ctx *gin.Context) {
    modelID := ctx.Param("id")

    if modelID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model ID is required"})
        return
    }

    query := `
        SELECT 
            m.id, m.user_id, m.first_name, m.last_name, m.username, m.email, 
            m.whatsapp, m.date_of_birth, m.gender, m.nationality, m.street, 
            m.city, m.residence_country, m.status, m.registration_step, m.deleted,
            m.created_at, m.updated_at,
            mm.experience, mm.height, mm.weight, mm.hips, mm.waist, 
            mm.hair_color, mm.eye_color, mm.photo, mm.additional_photos,
            md.document_issuer_country, md.document_type, md.document_front, md.document_back,
            mic.selfie_with_id, mic.verified as identity_verified,
            u.fullname as user_fullname, u.email as user_email, u.phone_number as user_phone
        FROM models m
        LEFT JOIN model_measurements mm ON m.id = mm.model_id
        LEFT JOIN model_documents md ON m.id = md.model_id
        LEFT JOIN model_identity_check mic ON m.id = mic.model_id
        LEFT JOIN users u ON m.user_id = u.userid
        WHERE m.id = $1 AND m.deleted = FALSE
    `

    var (
        id, userID, firstName, lastName, username, email, whatsapp string
        dateOfBirth, gender, nationality, street, city, residenceCountry, status string
        createdAt, updatedAt time.Time
        registrationStep int
        deleted bool
        experience, hairColor, eyeColor, photo, additionalPhotos sql.NullString
        height, weight, hips, waist sql.NullInt64
        docIssuerCountry, docType, docFront, docBack sql.NullString
        selfieWithID sql.NullString
        identityVerified sql.NullBool
        userFullname, userEmail, userPhone sql.NullString
    )

    err := database.DB.QueryRow(query, modelID).Scan(
        &id, &userID, &firstName, &lastName, &username, &email, &whatsapp,
        &dateOfBirth, &gender, &nationality, &street, &city, &residenceCountry, &status,
        &registrationStep, &deleted, &createdAt, &updatedAt, &experience, &height, &weight, 
        &hips, &waist, &hairColor, &eyeColor, &photo, &additionalPhotos, &docIssuerCountry,
        &docType, &docFront, &docBack, &selfieWithID, &identityVerified,
        &userFullname, &userEmail, &userPhone,
    )

    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
        return
    }

    model := gin.H{
        "id": id,
        "user_id": userID,
        "first_name": firstName,
        "last_name": lastName,
        "username": username,
        "email": email,
        "whatsapp": whatsapp,
        "date_of_birth": dateOfBirth,
        "gender": gender,
        "nationality": nationality,
        "street": street,
        "city": city,
        "residence_country": residenceCountry,
        "status": status,
        "registration_step": registrationStep,
        "deleted": deleted,
        "created_at": createdAt,
        "updated_at": updatedAt,
        "user_info": gin.H{
            "fullname": userFullname.String,
            "email": userEmail.String,
            "phone": userPhone.String,
        },
        "measurements": gin.H{
            "experience": experience.String,
            "height": height.Int64,
            "weight": weight.Int64,
            "hips": hips.Int64,
            "waist": waist.Int64,
            "hair_color": hairColor.String,
            "eye_color": eyeColor.String,
            "photo": photo.String,
            "additional_photos": additionalPhotos.String,
        },
        "documents": gin.H{
            "document_issuer_country": docIssuerCountry.String,
            "document_type": docType.String,
            "document_front": docFront.String,
            "document_back": docBack.String,
        },
        "identity_check": gin.H{
            "selfie_with_id": selfieWithID.String,
            "verified": identityVerified.Bool,
        },
    }

    ctx.JSON(http.StatusOK, gin.H{
        "model": model,
    })
}
