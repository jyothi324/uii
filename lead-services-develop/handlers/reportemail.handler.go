package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/utils"
	"github.com/tuliporg/agent-management-api/middleware"

)

func (h *Handler) SendReportEmail(c *gin.Context) {
	emailRaw := c.PostForm("email")
	subject := c.PostForm("subject")
	message := c.PostForm("message")

	if emailRaw == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Code:    "VALIDATION_ERROR",
			Details: "Email is required",
		})
		return
	}

	form, err := c.MultipartForm()
	if err != nil && err != http.ErrNotMultipart && err != http.ErrMissingBoundary {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Code:    "VALIDATION_ERROR",
			Details: "Failed to parse form: " + err.Error(),
		})
		return
	}

	var attachments []utils.EmailAttachment

	if form != nil {
		fileHeaders := form.File["files"] // matches frontend's formData.append("files", ...)
		for _, fh := range fileHeaders {
			f, err := fh.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to read file",
					Code:    "FILE_READ_ERROR",
					Details: err.Error(),
				})
				return
			}
			b, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to read file",
					Code:    "FILE_READ_ERROR",
					Details: err.Error(),
				})
				return
			}
			attachments = append(attachments, utils.EmailAttachment{Name: fh.Filename, Bytes: b})
		}
	}

	rawRecipients := strings.Split(emailRaw, ",")
	recipients := make([]string, 0, len(rawRecipients))
	for _, r := range rawRecipients {
		trimmed := strings.TrimSpace(r)
		if trimmed != "" {
			recipients = append(recipients, trimmed)
		}
	}
	if len(recipients) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Code:    "VALIDATION_ERROR",
			Details: "At least one valid email address is required",
		})
		return
	}

	if subject == "" {
		subject = "Report"
	}
	if message == "" {
		message = "Please find the attached report."
	}
	
	tenant := middleware.GetTenant(c)

	for _, to := range recipients {
		sendErr := utils.SendEmailWithAttachments(tenant,to, subject, message, attachments)
		if sendErr != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to send email",
				Code:    "EMAIL_ERROR",
				Details: "Failed sending to " + to + ": " + sendErr.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Report emailed successfully",
		"recipients": recipients,
	})
}
