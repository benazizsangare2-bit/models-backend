package handlers

import (
	"database/sql"
	//"encoding/json"
	"fmt"
	"models/database"
	"net/http"
	"os"

	//"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type HostessExperience struct {
    ID                string         `json:"id"`
    HostessID         string         `json:"hostess_id"`
    WorkExperience    string         `json:"work_experience"`
    Languages         pq.StringArray `json:"languages"`
    Skills            pq.StringArray `json:"skills"`
    Availability      string         `json:"availability"`
    PreferredEvents   pq.StringArray `json:"preferred_events"`
    PreviousWork      string         `json:"previous_hostess_work"`
    ReferenceContact  string         `json:"reference_contact"`
    Height            string         `json:"height"`
    Weight            string         `json:"weight"`
    HairColor         string         `json:"hair_color"`
    EyeColor          string         `json:"eye_color"`
    Photo             string         `json:"photo"`
    SocialInstagram   string         `json:"social_instagram"`
    SocialFacebook    string         `json:"social_facebook"`
    SocialTwitter     string         `json:"social_twitter"`
    SocialLinkedin    string         `json:"social_linkedin"`
    CreatedAt         string         `json:"created_at"`
    UpdatedAt         string         `json:"updated_at"`
}


// Create Hostess
func CreateHostess(ctx *gin.Context) {
	var req struct {
		FirstName        string `json:"first_name"`
		LastName         string `json:"last_name"`
		Username         string `json:"username"`
		Email            string `json:"email"`
		Whatsapp         string `json:"whatsapp"`
		DateOfBirth      string `json:"date_of_birth"`
		Gender           string `json:"gender"`
		Nationality      string `json:"nationality"`
		Street           string `json:"street"`
		City             string `json:"city"`
		ResidenceCountry string `json:"residence_country"`
        EmergencyName    string `json:"emergency_contact_name"`
        EmergencyRel     string `json:"emergency_contact_relationship"`
        EmergencyPhone   string `json:"emergency_contact_phone"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID := int(ctx.MustGet("user_id").(float64))
	email := ctx.MustGet("email").(string)

	// Validate DOB format
	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, expected YYYY-MM-DD"})
		return
	}

    // Insert into DB
    query := `
        INSERT INTO hostesses (
            user_id, first_name, last_name, username, email, whatsapp,
            date_of_birth, gender, nationality, street, city, residence_country,
            emergency_contact_name, emergency_contact_relationship, emergency_contact_phone
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
        RETURNING id
    `
	var hostessID string
	err = database.DB.QueryRow(query,
        userID, req.FirstName, req.LastName, req.Username, email,
        req.Whatsapp, dob, req.Gender, req.Nationality, req.Street, req.City, req.ResidenceCountry,
        req.EmergencyName, req.EmergencyRel, req.EmergencyPhone,
	).Scan(&hostessID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit the first step. Verify that all fields are correct and try again"})
		fmt.Println("DB error:", err)
		return
	}

	// ✅ Return ID so frontend can continue
	ctx.JSON(http.StatusOK, gin.H{
		"id":      hostessID,
		"message": "Step 1 saved successfully",
	})
}


// Add hostess experience and Skills
func AddHostessExperience(ctx *gin.Context) {
    hostessID := ctx.PostForm("hostess_id")
    if hostessID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Hostess ID is required"})
        return
    }

    // Parse fields
    workExperience := ctx.PostForm("work_experience")
    languages := strings.Split(strings.TrimSpace(ctx.PostForm("languages")), ",")
    skills := strings.Split(strings.TrimSpace(ctx.PostForm("skills")), ",")
    availability := ctx.PostForm("availability")
    preferredEvents := strings.Split(strings.TrimSpace(ctx.PostForm("preferred_events")), ",")
    previousHostessWork := ctx.PostForm("previous_hostess_work")
    referencesText := ctx.PostForm("reference_contact")

    // Optional physical attributes
    height := ctx.PostForm("height")
    weight := ctx.PostForm("weight")
    hairColor := ctx.PostForm("hair_color")
    eyeColor := ctx.PostForm("eye_color")

    // Optional socials
    socialInstagram := ctx.PostForm("social_instagram")
    socialFacebook := ctx.PostForm("social_facebook")
    socialTwitter := ctx.PostForm("social_twitter")
    socialLinkedin := ctx.PostForm("social_linkedin")

    // Ensure upload folder exists
    uploadDir := "uploads/hostesses"
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
        return
    }
    // Handle additional photos (array, up to 10)
    // ---- Handle additional photos (max 5) ----
    var photoPath []string
    form, err := ctx.MultipartForm()
    if err == nil && form.File["photo"] != nil {
        photos := form.File["photo"]
        if len(photos) < 5 {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": "Minimum 5 photos is required from the model"})
            return
        }

        for _, file := range photos {
            path := fmt.Sprintf("%s/%d_%s", uploadDir, time.Now().UnixNano(), file.Filename)
            if err := ctx.SaveUploadedFile(file, path); err != nil {
                ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to photo"})
                return
            }
            photoPath = append(photoPath, path)
        }
    } else {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "No photos where detected. Photos are required"})
        return
    }

    query := `
        INSERT INTO hostess_experience (
            hostess_id, work_experience, languages, skills, availability,
            preferred_events, previous_hostess_work, reference_contact,
            height, weight, hair_color, eye_color,
            photo,
            social_instagram, social_facebook, social_twitter, social_linkedin
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
    `
    _, err = database.DB.Exec(query,
        hostessID, workExperience, pq.Array(languages), pq.Array(skills), availability,
        pq.Array(preferredEvents), previousHostessWork, referencesText,
        height, weight, hairColor, eyeColor,
        pq.Array(photoPath),
        socialInstagram, socialFacebook, socialTwitter, socialLinkedin,
    )

    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save experience"})
        fmt.Println("DB error:", err)
        return
    }

    // ✅ Optionally update registration step
    _, _ = database.DB.Exec(`UPDATE hostesses SET registration_step = 2 WHERE id = $1`, hostessID)

    ctx.JSON(http.StatusOK, gin.H{"message": "Step 2 saved successfully"})
}

/// User's last progress for Hostess registration
func GetHostessProgress(ctx *gin.Context) {
    userID := int(ctx.MustGet("user_id").(float64))

    var hostessID string
    var step int
    err := database.DB.QueryRow(`
        SELECT id, registration_step
        FROM hostesses
        WHERE user_id = $1 AND deleted = FALSE
        ORDER BY created_at DESC
        LIMIT 1
    `, userID).Scan(&hostessID, &step)

    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "No hostess found for user"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "hostess_id": hostessID,
        "current_step": step,
    })
}

// Admin endpoint to get all hostesses with complete information for review
func AdminGetAllHostesses(ctx *gin.Context) {
    // Get query parameters for filtering
    status := ctx.Query("status") // pending, under_review, approved, rejected
    page := ctx.DefaultQuery("page", "1")
    limit := ctx.DefaultQuery("limit", "10")

    // Build base query
    baseQuery := `
        SELECT 
            h.id, h.user_id, h.first_name, h.last_name, h.username, h.email, 
            h.whatsapp, h.date_of_birth, h.gender, h.nationality, h.street, 
            h.city, h.residence_country, h.status, h.registration_step, h.deleted,
            h.emergency_contact_name, h.emergency_contact_relationship, h.emergency_contact_phone,
            h.created_at, h.updated_at,
            he.work_experience, he.languages, he.skills, he.availability, he.preferred_events,
            he.previous_hostess_work, he.reference_contact, he.height, he.weight, 
            he.hair_color, he.eye_color, he.photo,
            he.social_instagram, he.social_facebook, he.social_twitter, he.social_linkedin,
            hd.document_issuer_country, hd.document_type, hd.document_front, hd.document_back,
            hic.selfie_with_id, hic.verified as identity_verified,
            u.fullname as user_fullname, u.email as user_email, u.phone_number as user_phone
        FROM hostesses h
        LEFT JOIN hostess_experience he ON h.id = he.hostess_id
        LEFT JOIN hostess_documents hd ON h.id = hd.hostess_id
        LEFT JOIN hostess_identity_check hic ON h.id = hic.hostess_id
        LEFT JOIN users u ON h.user_id = u.userid
        WHERE h.deleted = FALSE
    `

    // Add status filter if provided
    if status != "" {
        baseQuery += " AND h.status = $1"
    }
    
    baseQuery += " ORDER BY h.created_at DESC"

    // Execute query
    var rows *sql.Rows
    var err error
    
    if status != "" {
        rows, err = database.DB.Query(baseQuery, status)
    } else {
        rows, err = database.DB.Query(baseQuery)
    }

    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch hostesses"})
        fmt.Println("Database query error:", err)
        return
    }
    defer rows.Close()

    var hostesses []gin.H
    for rows.Next() {
        var (
            id, userID, firstName, lastName, username, email, whatsapp string
            dateOfBirth, gender, nationality, street, city, residenceCountry, status string
            createdAt, updatedAt time.Time
            registrationStep int
            deleted bool
            emergencyName, emergencyRel, emergencyPhone sql.NullString
            workExperience, availability, previousWork, referenceContact sql.NullString
            height, weight, hairColor, eyeColor, photo sql.NullString
            socialInstagram, socialFacebook, socialTwitter, socialLinkedin sql.NullString
            docIssuerCountry, docType, docFront, docBack sql.NullString
            selfieWithID sql.NullString
            identityVerified sql.NullBool
            userFullname, userEmail, userPhone sql.NullString
            languages, skills, preferredEvents pq.StringArray
        )

        err := rows.Scan(
            &id, &userID, &firstName, &lastName, &username, &email, &whatsapp,
            &dateOfBirth, &gender, &nationality, &street, &city, &residenceCountry, &status,
            &registrationStep, &deleted, &emergencyName, &emergencyRel, &emergencyPhone,
            &createdAt, &updatedAt, &workExperience, &languages, &skills, &availability,
            &preferredEvents, &previousWork, &referenceContact, &height, &weight,
            &hairColor, &eyeColor, &photo, &socialInstagram,
            &socialFacebook, &socialTwitter, &socialLinkedin, &docIssuerCountry,
            &docType, &docFront, &docBack, &selfieWithID, &identityVerified,
            &userFullname, &userEmail, &userPhone,
        )
        if err != nil {
            fmt.Println("Row scan error:", err)
            continue
        }

        hostess := gin.H{
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
            "emergency_contact": gin.H{
                "name": emergencyName.String,
                "relationship": emergencyRel.String,
                "phone": emergencyPhone.String,
            },
            "user_info": gin.H{
                "fullname": userFullname.String,
                "email": userEmail.String,
                "phone": userPhone.String,
            },
            "experience": gin.H{
                "work_experience": workExperience.String,
                "languages": languages,
                "skills": skills,
                "availability": availability.String,
                "preferred_events": preferredEvents,
                "previous_hostess_work": previousWork.String,
                "reference_contact": referenceContact.String,
                "height": height.String,
                "weight": weight.String,
                "hair_color": hairColor.String,
                "eye_color": eyeColor.String,
                "photo": photo.String,
                "social_instagram": socialInstagram.String,
                "social_facebook": socialFacebook.String,
                "social_twitter": socialTwitter.String,
                "social_linkedin": socialLinkedin.String,
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
        hostesses = append(hostesses, hostess)
    }

    // Get total count for pagination
    countQuery := "SELECT COUNT(*) FROM hostesses WHERE deleted = FALSE"
    var totalCount int
    if status != "" {
        err = database.DB.QueryRow(countQuery + " AND status = $1", status).Scan(&totalCount)
    } else {
        err = database.DB.QueryRow(countQuery).Scan(&totalCount)
    }

    if err != nil {
        fmt.Println("Count query error:", err)
        totalCount = len(hostesses)
    }

    ctx.JSON(http.StatusOK, gin.H{
        "hostesses": hostesses,
        "total_count": totalCount,
        "filtered_count": len(hostesses),
        "status_filter": status,
        "page": page,
        "limit": limit,
    })
}

// Admin endpoint to get a specific hostess by ID with complete information
func AdminGetHostessById(ctx *gin.Context) {
    hostessID := ctx.Param("id")

    if hostessID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Hostess ID is required"})
        return
    }

    query := `
        SELECT 
            h.id, h.user_id, h.first_name, h.last_name, h.username, h.email, 
            h.whatsapp, h.date_of_birth, h.gender, h.nationality, h.street, 
            h.city, h.residence_country, h.status, h.registration_step, h.deleted,
            h.emergency_contact_name, h.emergency_contact_relationship, h.emergency_contact_phone,
            h.created_at, h.updated_at,
            he.work_experience, he.languages, he.skills, he.availability, he.preferred_events,
            he.previous_hostess_work, he.reference_contact, he.height, he.weight, 
            he.hair_color, he.eye_color, he.photo,
            he.social_instagram, he.social_facebook, he.social_twitter, he.social_linkedin,
            hd.document_issuer_country, hd.document_type, hd.document_front, hd.document_back,
            hic.selfie_with_id, hic.verified as identity_verified,
            u.fullname as user_fullname, u.email as user_email, u.phone_number as user_phone
        FROM hostesses h
        LEFT JOIN hostess_experience he ON h.id = he.hostess_id
        LEFT JOIN hostess_documents hd ON h.id = hd.hostess_id
        LEFT JOIN hostess_identity_check hic ON h.id = hic.hostess_id
        LEFT JOIN users u ON h.user_id = u.userid
        WHERE h.id = $1 AND h.deleted = FALSE
    `

    var (
        id, userID, firstName, lastName, username, email, whatsapp string
        dateOfBirth, gender, nationality, street, city, residenceCountry, status string
        createdAt, updatedAt time.Time
        registrationStep int
        deleted bool
        emergencyName, emergencyRel, emergencyPhone sql.NullString
        workExperience, availability, previousWork, referenceContact sql.NullString
        height, weight, hairColor, eyeColor, photo sql.NullString
        socialInstagram, socialFacebook, socialTwitter, socialLinkedin sql.NullString
        docIssuerCountry, docType, docFront, docBack sql.NullString
        selfieWithID sql.NullString
        identityVerified sql.NullBool
        userFullname, userEmail, userPhone sql.NullString
        languages, skills, preferredEvents pq.StringArray
    )

    err := database.DB.QueryRow(query, hostessID).Scan(
        &id, &userID, &firstName, &lastName, &username, &email, &whatsapp,
        &dateOfBirth, &gender, &nationality, &street, &city, &residenceCountry, &status,
        &registrationStep, &deleted, &emergencyName, &emergencyRel, &emergencyPhone,
        &createdAt, &updatedAt, &workExperience, &languages, &skills, &availability,
        &preferredEvents, &previousWork, &referenceContact, &height, &weight,
        &hairColor, &eyeColor, &photo, &socialInstagram,
        &socialFacebook, &socialTwitter, &socialLinkedin, &docIssuerCountry,
        &docType, &docFront, &docBack, &selfieWithID, &identityVerified,
        &userFullname, &userEmail, &userPhone,
    )

    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Hostess not found"})
        return
    }

    hostess := gin.H{
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
        "emergency_contact": gin.H{
            "name": emergencyName.String,
            "relationship": emergencyRel.String,
            "phone": emergencyPhone.String,
        },
        "user_info": gin.H{
            "fullname": userFullname.String,
            "email": userEmail.String,
            "phone": userPhone.String,
        },
        "experience": gin.H{
            "work_experience": workExperience.String,
            "languages": languages,
            "skills": skills,
            "availability": availability.String,
            "preferred_events": preferredEvents,
            "previous_hostess_work": previousWork.String,
            "reference_contact": referenceContact.String,
            "height": height.String,
            "weight": weight.String,
            "hair_color": hairColor.String,
            "eye_color": eyeColor.String,
            "photo": photo.String,
            "social_instagram": socialInstagram.String,
            "social_facebook": socialFacebook.String,
            "social_twitter": socialTwitter.String,
            "social_linkedin": socialLinkedin.String,
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
        "hostess": hostess,
    })
}

// Admin endpoint to approve or reject a hostess
// Separate handler for approve hostess
func AdminApproveHostess(ctx *gin.Context) {
    hostessID := ctx.Param("id")
    handleHostessStatusUpdate(ctx, hostessID, "approved")
}

// Separate handler for reject hostess  
func AdminRejectHostess(ctx *gin.Context) {
    hostessID := ctx.Param("id")
    handleHostessStatusUpdate(ctx, hostessID, "rejected")
}

// Common function for hostess status update
func handleHostessStatusUpdate(ctx *gin.Context, hostessID string, newStatus string) {
    if hostessID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Hostess ID is required"})
        return
    }

    // Verify the hostess exists and is not deleted
    var currentStatus string
    err := database.DB.QueryRow(`
        SELECT status FROM hostesses 
        WHERE id = $1 AND deleted = FALSE
    `, hostessID).Scan(&currentStatus)

    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Hostess not found"})
        return
    }

    // Check if hostess is already in final state
    if currentStatus == "approved" || currentStatus == "rejected" {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": fmt.Sprintf("Hostess is already %s", currentStatus),
        })
        return
    }

    // Parse request body for admin notes (optional)
    var req struct {
        AdminNotes string `json:"admin_notes"`
    }
    if err := ctx.ShouldBindJSON(&req); err != nil {
        // It's okay if no JSON body is provided
        req.AdminNotes = ""
    }

    // Update hostess status
    query := `
        UPDATE hostesses 
        SET status = $1, updated_at = NOW() 
        WHERE id = $2
    `
    
    _, err = database.DB.Exec(query, newStatus, hostessID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update hostess status"})
        fmt.Println("Update error:", err)
        return
    }

    // Log the action
    logMessage := fmt.Sprintf("Hostess %s %s by admin", hostessID, newStatus)
    if req.AdminNotes != "" {
        logMessage += fmt.Sprintf(" - Notes: %s", req.AdminNotes)
    }
    fmt.Println(logMessage)

    ctx.JSON(http.StatusOK, gin.H{
        "message": fmt.Sprintf("Hostess %s successfully", newStatus),
        "hostess_id": hostessID,
        "new_status": newStatus,
        "admin_notes": req.AdminNotes,
    })
}

// Soft delete hostess by setting deleted = true
func DeleteHostess(ctx *gin.Context) {
    userID := int(ctx.MustGet("user_id").(float64))
    hostessID := ctx.Param("id")

    if hostessID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Hostess ID is required"})
        return
    }

    // Verify the hostess belongs to the user
    var ownerID int
    err := database.DB.QueryRow(`
        SELECT user_id FROM hostesses 
        WHERE id = $1 AND deleted = FALSE
    `, hostessID).Scan(&ownerID)

    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Hostess not found"})
        return
    }

    if ownerID != userID {
        ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own hostess"})
        return
    }

    // Soft delete the hostess
    _, err = database.DB.Exec(`
        UPDATE hostesses SET deleted = TRUE, updated_at = NOW() 
        WHERE id = $1
    `, hostessID)

    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete hostess"})
        fmt.Println("Delete error:", err)
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "message": "Hostess deleted successfully",
    })
}

// Update hostess information
func UpdateHostess(ctx *gin.Context) {
    userID := int(ctx.MustGet("user_id").(float64))
    hostessID := ctx.Param("id")

    if hostessID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Hostess ID is required"})
        return
    }

    // Verify the hostess belongs to the user
    var ownerID int
    err := database.DB.QueryRow(`
        SELECT user_id FROM hostesses 
        WHERE id = $1 AND deleted = FALSE
    `, hostessID).Scan(&ownerID)

    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Hostess not found"})
        return
    }

    if ownerID != userID {
        ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own hostess"})
        return
    }

    // Parse request body
    var req struct {
        FirstName        string `json:"first_name"`
        LastName         string `json:"last_name"`
        Username         string `json:"username"`
        Email            string `json:"email"`
        Whatsapp         string `json:"whatsapp"`
        DateOfBirth      string `json:"date_of_birth"`
        Gender           string `json:"gender"`
        Nationality      string `json:"nationality"`
        Street           string `json:"street"`
        City             string `json:"city"`
        ResidenceCountry string `json:"residence_country"`
        EmergencyName    string `json:"emergency_contact_name"`
        EmergencyRel     string `json:"emergency_contact_relationship"`
        EmergencyPhone   string `json:"emergency_contact_phone"`
    }

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

    // Update hostess information
    query := `
        UPDATE hostesses SET 
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
            emergency_contact_name = COALESCE($12, emergency_contact_name),
            emergency_contact_relationship = COALESCE($13, emergency_contact_relationship),
            emergency_contact_phone = COALESCE($14, emergency_contact_phone),
            updated_at = NOW()
        WHERE id = $1
    `

    _, err = database.DB.Exec(query,
        hostessID,
        req.FirstName, req.LastName, req.Username, req.Whatsapp,
        dob, req.Gender, req.Nationality, req.Street, req.City, req.ResidenceCountry,
        req.EmergencyName, req.EmergencyRel, req.EmergencyPhone,
    )

    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update hostess"})
        fmt.Println("Update error:", err)
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "message": "Hostess updated successfully",
        "hostess_id": hostessID,
    })
}

/// Document verification front and back of the document 
func AddHostessDocuments(ctx *gin.Context) {
    hostessID := ctx.PostForm("hostess_id")
    if hostessID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Hostess ID is required"})
        return
    }

    documentIssuerCountry := ctx.PostForm("documentIssuerCountry")
    documentType := ctx.PostForm("documentType")

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

    uploadDir := "uploads/hostesses/documents"
    os.MkdirAll(uploadDir, os.ModePerm)

    frontPath := fmt.Sprintf("%s/%s_%d_%s", uploadDir, hostessID, time.Now().Unix(), frontFile.Filename)
    backPath := fmt.Sprintf("%s/%s_%d_%s", uploadDir, hostessID, time.Now().Unix(), backFile.Filename)

    if err := ctx.SaveUploadedFile(frontFile, frontPath); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save front image"})
        return
    }

    if err := ctx.SaveUploadedFile(backFile, backPath); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save back image"})
        return
    }

    query := `
        INSERT INTO hostess_documents (hostess_id, document_issuer_country, document_type, document_front, document_back)
        VALUES ($1, $2, $3, $4, $5)
    `
    _, err = database.DB.Exec(query, hostessID, documentIssuerCountry, documentType, frontPath, backPath)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save documents"})
        return
    }

    // Optional: advance registration step
    _, _ = database.DB.Exec(`UPDATE hostesses SET registration_step = 3 WHERE id = $1`, hostessID)

    ctx.JSON(http.StatusOK, gin.H{"message": "✅ Documents uploaded successfully!"})
}


/// Upload your face with the document id 
func UploadHostessIdentityCheck(ctx *gin.Context) {
    hostessID := ctx.PostForm("hostess_id")
    if hostessID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Hostess ID is required"})
        return
    }

    file, err := ctx.FormFile("selfie_with_id")
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Selfie with ID is required"})
        return
    }

    uploadDir := "uploads/hostesses/identity"
    os.MkdirAll(uploadDir, os.ModePerm)

    filePath := fmt.Sprintf("%s/%s_%d_%s", uploadDir, hostessID, time.Now().Unix(), file.Filename)
    if err := ctx.SaveUploadedFile(file, filePath); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save selfie"})
        return
    }

    _, err = database.DB.Exec(`
        INSERT INTO hostess_identity_check (hostess_id, selfie_with_id)
        VALUES ($1, $2)
    `, hostessID, filePath)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database insert failed"})
        return
    }

    // Optional: mark registration as complete
    _, _ = database.DB.Exec(`UPDATE hostesses SET registration_step = 4, status = 'under_review' WHERE id = $1`, hostessID)

    ctx.JSON(http.StatusOK, gin.H{"message": "✅ Identity check submitted successfully!"})
}


// Get approved hostesses
func GetApprovedHostesses(ctx *gin.Context) {
    query := `
        SELECT 
            h.id, h.user_id, h.first_name, h.last_name, h.username, h.email, 
            h.whatsapp, h.date_of_birth, h.gender, h.nationality, h.street, 
            h.city, h.residence_country, h.status, h.registration_step, h.deleted,
            h.emergency_contact_name, h.emergency_contact_relationship, h.emergency_contact_phone,
            h.created_at, h.updated_at,

            he.work_experience, he.languages, he.skills, he.availability, he.preferred_events, 
            he.previous_hostess_work, he.reference_contact, 
            he.height, he.weight, he.hair_color, he.eye_color, he.photo,
            he.social_instagram, he.social_facebook, he.social_twitter, he.social_linkedin,

            hd.document_issuer_country, hd.document_type, hd.document_front, hd.document_back,
            hic.selfie_with_id, hic.verified as identity_verified

        FROM hostesses h
        LEFT JOIN hostess_experience he ON h.id = he.hostess_id
        LEFT JOIN hostess_documents hd ON h.id = hd.hostess_id
        LEFT JOIN hostess_identity_check hic ON h.id = hic.hostess_id
        WHERE h.status = 'approved' AND h.deleted = FALSE
        ORDER BY h.created_at DESC
    `
    rows, err := database.DB.Query(query)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approved hostesses"})
        fmt.Println("Database query error:", err)
        return
    }
    defer rows.Close()

    var hostesses []gin.H

    for rows.Next() {
        var (
            id, userID, firstName, lastName, username, email, whatsapp string
            dateOfBirth, gender, nationality, street, city, residenceCountry, status string
            createdAt, updatedAt time.Time
            registrationStep int
            deleted bool

            emergencyName, emergencyRelationship, emergencyPhone sql.NullString

            workExperience, availability, previousWork, referenceContact sql.NullString
            height, weight, hairColor, eyeColor sql.NullString
            socialInstagram, socialFacebook, socialTwitter, socialLinkedin sql.NullString
            photo sql.NullString
            languages, skills, preferredEvents pq.StringArray

            docIssuerCountry, docType, docFront, docBack sql.NullString
            selfieWithID sql.NullString
            identityVerified sql.NullBool
        )

        err := rows.Scan(
            &id, &userID, &firstName, &lastName, &username, &email, &whatsapp,
            &dateOfBirth, &gender, &nationality, &street, &city, &residenceCountry,
            &status, &registrationStep, &deleted,
            &emergencyName, &emergencyRelationship, &emergencyPhone,
            &createdAt, &updatedAt,

            &workExperience, &languages, &skills, &availability, &preferredEvents,
            &previousWork, &referenceContact,
            &height, &weight, &hairColor, &eyeColor, &photo,
            &socialInstagram, &socialFacebook, &socialTwitter, &socialLinkedin,

            &docIssuerCountry, &docType, &docFront, &docBack,
            &selfieWithID, &identityVerified,
        )
        if err != nil {
            fmt.Println("Row scan error:", err)
            continue
        }

        // Helper functions
        getStringValue := func(value sql.NullString, defaultValue string) string {
            if value.Valid && value.String != "" {
                return value.String
            }
            return defaultValue
        }
        
        getBoolValue := func(value sql.NullBool, defaultValue bool) bool {
            if value.Valid {
                return value.Bool
            }
            return defaultValue
        }

        // Parse photos into array (similar to models)
        var photoArray []string
        if photo.Valid && photo.String != "" {
            photoStr := photo.String
            
            // Handle PostgreSQL array format: {photo1.jpg,photo2.jpg,photo3.jpg}
            if strings.HasPrefix(photoStr, "{") && strings.HasSuffix(photoStr, "}") {
                cleaned := photoStr[1:len(photoStr)-1] // Remove curly braces
                photos := strings.Split(cleaned, ",")
                for _, p := range photos {
                    cleanedPhoto := strings.TrimSpace(p)
                    cleanedPhoto = strings.Trim(cleanedPhoto, "\"")
                    if cleanedPhoto != "" {
                        photoArray = append(photoArray, cleanedPhoto)
                    }
                }
            } else if strings.Contains(photoStr, ",") {
                // Handle comma-separated string
                photos := strings.Split(photoStr, ",")
                for _, p := range photos {
                    cleanedPhoto := strings.TrimSpace(p)
                    if cleanedPhoto != "" {
                        photoArray = append(photoArray, cleanedPhoto)
                    }
                }
            } else {
                // Single photo
                photoArray = []string{strings.TrimSpace(photoStr)}
            }
        }

        // Get first photo for main field
        var firstPhoto string
        if len(photoArray) > 0 {
            firstPhoto = photoArray[0]
        } else {
            firstPhoto = "/uploads/default.jpg"
        }

        hostess := gin.H{
            "id":                id,
            "user_id":           userID,
            "first_name":        firstName,
            "last_name":         lastName,
            "username":          username,
            "email":             email,
            "whatsapp":          whatsapp,
            "date_of_birth":     dateOfBirth,
            "gender":            gender,
            "nationality":       nationality,
            "street":            street,
            "city":              city,
            "residence_country": residenceCountry,
            "status":            status,
            "registration_step": registrationStep,
            "deleted":           deleted,
            "created_at":        createdAt,
            "updated_at":        updatedAt,

            "emergency_contact": gin.H{
                "name":         getStringValue(emergencyName, ""),
                "relationship": getStringValue(emergencyRelationship, ""),
                "phone":        getStringValue(emergencyPhone, ""),
            },

            "experience": gin.H{
                "work_experience":        getStringValue(workExperience, ""),
                "languages":              languages,
                "skills":                 skills,
                "availability":           getStringValue(availability, ""),
                "preferred_events":       preferredEvents,
                "previous_hostess_work":  getStringValue(previousWork, ""),
                "reference_contact":      getStringValue(referenceContact, ""),
                "height":                 getStringValue(height, ""),
                "weight":                 getStringValue(weight, ""),
                "hair_color":             getStringValue(hairColor, ""),
                "eye_color":              getStringValue(eyeColor, ""),
                "photo":                  photoArray, // All photos as array
                "social_instagram":       getStringValue(socialInstagram, ""),
                "social_facebook":        getStringValue(socialFacebook, ""),
                "social_twitter":         getStringValue(socialTwitter, ""),
                "social_linkedin":        getStringValue(socialLinkedin, ""),
            },

            "documents": gin.H{
                "document_issuer_country": getStringValue(docIssuerCountry, ""),
                "document_type":           getStringValue(docType, ""),
                "document_front":          getStringValue(docFront, ""),
                "document_back":           getStringValue(docBack, ""),
            },

            "identity_check": gin.H{
                "selfie_with_id": getStringValue(selfieWithID, ""),
                "verified":       getBoolValue(identityVerified, false),
            },

            "photo": firstPhoto, // Only the first photo for the main field
        }

        hostesses = append(hostesses, hostess)
    }

    ctx.JSON(http.StatusOK, gin.H{
        "hostesses": hostesses,
        "count":     len(hostesses),
    })
}


// AdminUpdateHostess updates hostess information
func AdminUpdateHostess(ctx *gin.Context) {
    hostessID := ctx.Param("id")
    handleHostessUpdate(ctx, hostessID)
}

// Common function for hostess update
func handleHostessUpdate(ctx *gin.Context, hostessID string) {
    if hostessID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Hostess ID is required"})
        return
    }

    // Verify the hostess exists and is not deleted
    var exists bool
    err := database.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM hostesses WHERE id = $1 AND deleted = FALSE)
    `, hostessID).Scan(&exists)

    if err != nil || !exists {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Hostess not found"})
        return
    }

    // Parse request body
    var req struct {
        FirstName         string `json:"first_name"`
        LastName          string `json:"last_name"`
        Username          string `json:"username"`
        Email             string `json:"email"`
        WhatsApp          string `json:"whatsapp"`
        DateOfBirth       string `json:"date_of_birth"`
        Gender            string `json:"gender"`
        Nationality       string `json:"nationality"`
        Street            string `json:"street"`
        City              string `json:"city"`
        ResidenceCountry  string `json:"residence_country"`
        Status            string `json:"status"`
        
        Measurements struct {
            Experience     string `json:"experience"`
            Height         int    `json:"height"`
            Weight         int    `json:"weight"`
            Hips           int    `json:"hips"`
            Waist          int    `json:"waist"`
            HairColor      string `json:"hair_color"`
            EyeColor       string `json:"eye_color"`
            SocialInstagram string `json:"social_instagram"`
            SocialFacebook  string `json:"social_facebook"`
            SocialTwitter   string `json:"social_twitter"`
            SocialLinkedin  string `json:"social_linkedin"`
        } `json:"measurements"`
        
        Documents struct {
            DocumentIssuerCountry string `json:"document_issuer_country"`
            DocumentType         string `json:"document_type"`
        } `json:"documents"`
    }

    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
        return
    }

    // Start transaction
    tx, err := database.DB.Begin()
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
        return
    }
    defer tx.Rollback()

    // Update main hostess table
    _, err = tx.Exec(`
        UPDATE hostesses 
        SET first_name = $1, last_name = $2, username = $3, email = $4, 
            whatsapp = $5, date_of_birth = $6, gender = $7, nationality = $8,
            street = $9, city = $10, residence_country = $11, status = $12,
            updated_at = NOW()
        WHERE id = $13
    `, req.FirstName, req.LastName, req.Username, req.Email, req.WhatsApp,
        req.DateOfBirth, req.Gender, req.Nationality, req.Street, req.City,
        req.ResidenceCountry, req.Status, hostessID)

    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update hostess basic info"})
        return
    }

    // Update measurements
    _, err = tx.Exec(`
        UPDATE hostess_measurements 
        SET experience = $1, height = $2, weight = $3, hips = $4, waist = $5,
            hair_color = $6, eye_color = $7, social_instagram = $8, 
            social_facebook = $9, social_twitter = $10, social_linkedin = $11,
            updated_at = NOW()
        WHERE hostess_id = $12
    `, req.Measurements.Experience, req.Measurements.Height, req.Measurements.Weight,
        req.Measurements.Hips, req.Measurements.Waist, req.Measurements.HairColor,
        req.Measurements.EyeColor, req.Measurements.SocialInstagram,
        req.Measurements.SocialFacebook, req.Measurements.SocialTwitter,
        req.Measurements.SocialLinkedin, hostessID)

    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update hostess measurements"})
        return
    }

    // Update documents
    _, err = tx.Exec(`
        UPDATE hostess_documents 
        SET document_issuer_country = $1, document_type = $2, updated_at = NOW()
        WHERE hostess_id = $3
    `, req.Documents.DocumentIssuerCountry, req.Documents.DocumentType, hostessID)

    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update hostess documents"})
        return
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit changes"})
        return
    }

    // Log the action
    fmt.Printf("Hostess %s updated by admin\n", hostessID)

    ctx.JSON(http.StatusOK, gin.H{
        "message":    "Hostess updated successfully",
        "hostess_id": hostessID,
    })
}