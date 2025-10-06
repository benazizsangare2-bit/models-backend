package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mailjet/mailjet-apiv3-go"
)

type ContactForm struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Company   string `json:"company"`
	EventType string `json:"eventType"`
	EventDate string `json:"eventDate"`
	Message   string `json:"message"`
}

// POST /api/contact
func HandleContact(ctx *gin.Context) {
	var form ContactForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
		})
		return
	}

	// ✅ Validate required fields
	if form.Name == "" || form.Email == "" || form.Message == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields: name, email, and message are required",
		})
		return
	}

	// ✅ Load Mailjet credentials from environment
	apiKey := os.Getenv("MAILJET_API_KEY")
	apiSecret := os.Getenv("MAILJET_API_SECRET")
	senderEmail := os.Getenv("MAILJET_SENDER_EMAIL")
	senderName := os.Getenv("MAILJET_SENDER_NAME")

	if apiKey == "" || apiSecret == "" || senderEmail == "" {
		fmt.Println("Error: Missing Mailjet environment variables")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Email service not configured properly. Missing environment variables.",
		})
		return
	}

	mailjetClient := mailjet.NewMailjetClient(apiKey, apiSecret)

	// ✅ Construct the email that goes to the agency
	subject := fmt.Sprintf("📩 New Contact Form Submission from %s", form.Name)
	body := fmt.Sprintf(`
You have received a new contact form submission.

--------------------------------------
Name: %s
Email: %s
Phone: %s
Company: %s
Event Type: %s
Event Date: %s

Message:
%s
--------------------------------------

This message was submitted from the website contact form.
`, form.Name, form.Email, form.Phone, form.Company, form.EventType, form.EventDate, form.Message)

	// ✅ Send email to your agency inbox (MAILJET_SENDER_EMAIL)
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: senderEmail, // From your agency email
				Name:  senderName,
			},
			To: &mailjet.RecipientsV31{
				{
					Email: senderEmail, // Send TO your agency inbox
					Name:  "Agency Inbox",
				},
			},
			Subject:  subject,
			TextPart: body,
			ReplyTo: &mailjet.RecipientV31{ // ✅ So you can reply directly to the user
				Email: form.Email,
				Name:  form.Name,
			},
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}
	response, err := mailjetClient.SendMailV31(&messages)
	if err != nil {
		fmt.Printf("Mailjet error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send email. Please try again later.",
			"details": err.Error(),
		})
		return
	}

	fmt.Printf("Contact email successfully sent to agency inbox! Response: %+v\n", response)

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Message sent successfully! We'll get back to you soon.",
	})
}
