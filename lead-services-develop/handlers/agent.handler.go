package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kr/pretty"
	_ "github.com/lib/pq"
	"github.com/pquerna/otp/totp"
	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"github.com/tuliporg/agent-management-api/utils"
)

type Handler struct {
	DB db.Database
}

func NewHandler() *Handler {
	return &Handler{}
}
func (h *Handler) CheckCompanyExists(c *gin.Context) {
	companyName := c.Query("company")

	if companyName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Company name required",
		})
		return
	}

	companyName = strings.TrimSpace(strings.ToLower(companyName))
	dbName := fmt.Sprintf("tenant_%s", companyName)

	if !checkDatabaseExists(dbName) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Company not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Company verified successfully",
		"company": companyName,
	})
}

func checkDatabaseExists(dbName string) bool {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		dbName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return false
	}
	defer db.Close()

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = db.QueryRow(query, dbName).Scan(&exists)
	pretty.Print(err, "emmmmm")
	return err == nil && exists
}

func grantAllPermissionsToManager(db db.Database, agentID string) {
	var permissions []models.Permissions
	if err := db.Find(&permissions); err != nil {
		log.Printf("MANAGER_PERMS: failed to fetch permissions for agent %s: %v", agentID, err)
		return
	}

	var agent models.Agent
	if err := db.Where("id = ?", agentID).First(&agent); err != nil {
		log.Printf("MANAGER_PERMS: failed to fetch agent %s: %v", agentID, err)
		return
	}

	for _, perm := range permissions {
		if err := db.Model(&perm).Association("User").Append(&agent); err != nil {
			log.Printf("MANAGER_PERMS: failed to add agent %s to perm %d (%s:%s): %v",
				agentID, perm.ID, perm.Source, perm.Action, err)
		}
	}

	log.Printf("MANAGER_PERMS: granted all %d permissions to manager %s", len(permissions), agentID)
}

func (h *Handler) CreateAgent(c *gin.Context) {

	var req models.CreateAgentRequest
	db := middleware.GetDB(c)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	// Validate required fields
	if req.Name == "" || req.Email == "" || req.Role == "" {

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required fields",
			Code:    "VALIDATION_ERROR",
			Details: "Name, Email, Role are required",
		})
		return
	}

	var existingAgent models.Agent
	if err := db.Where("email = ?", req.Email).First(&existingAgent); err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "Email already exists",
			Code:    "DUPLICATE_EMAIL",
			Details: "An agent with this email address already exists",
		})
		return
	}
	if req.Details.TOTPEnabled {
		if req.TOTPSecret == "" || req.TOTPCode == "" {
			totpSecret, totpURL, err := utils.GenerateTOTP(req.Email, true)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate TOTP"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message":     "Scan QR and submit TOTP code to complete registration",
				"totp_secret": totpSecret,
				"totp_url":    totpURL,
			})
			return
		}
		if !utils.ValidateTOTP(req.TOTPCode, req.TOTPSecret) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid TOTP code"})
			return
		}
		req.Details.TOTPSecret = req.TOTPSecret
	}

	req.ID = utils.GenerateEmployeeID(db)
	req.PasswordHashed, _ = utils.HashPassword(req.Password)
	token, err := middleware.GenerateJWT(req.ID, req.Name, req.Email, string(req.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	agent := req.Agent
	agent.ID = req.ID
	agent.PasswordHashed = req.PasswordHashed
	agent.Details = req.Details

	if strings.EqualFold(string(agent.Role), "MANAGER") {
		agent.ShowPrices = true
	}

	if err := db.Create(&agent); err != nil {
		return
	}

	// ── Grant all permissions automatically for MANAGER role ──
	if strings.EqualFold(string(agent.Role), "MANAGER") {
		go grantAllPermissionsToManager(db, agent.ID)
	}

	agent.PasswordHashed = ""
	agent.Password = ""
	response := gin.H{
		"message": "User registered successfully",
		"id":      agent.ID,
		"token":   token,
	}
	if agent.Details.TOTPEnabled {
		response["totp_secret"] = agent.Details.TOTPSecret
	}
	c.JSON(http.StatusCreated, response)
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	TOTPCode string `json:"totp_code"`
}

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func (h *Handler) UpdatePassword(c *gin.Context) {
	db := middleware.GetDB(c)
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}
	var agent models.Agent
	db.Where("email = ?", req.Email).First(&agent)
	hashedPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	agent.PasswordHashed = hashedPwd
	if err := db.Save(&agent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update password",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password updated successfully",
	})
}

func (h *Handler) Login(c *gin.Context) {
	var loginReq LoginRequest
	db := middleware.GetDB(c)
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !IsValidEmail(loginReq.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	if len(loginReq.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters long"})
		return
	}

	var agent models.Agent

	if err := db.Where("email = ?", loginReq.Email).First(&agent); err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if !utils.CheckPasswordHash(loginReq.Password, agent.PasswordHashed) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if agent.Details.TOTPEnabled {
		if loginReq.TOTPCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "TOTP code required", "require_totp": true})
			return
		}

		if !totp.Validate(loginReq.TOTPCode, agent.Details.TOTPSecret) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid TOTP code"})
			return
		}
	}

	token, err := middleware.GenerateJWT(agent.ID, agent.Name, agent.Email, string(agent.Role))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   int(middleware.TokenExpiry.Seconds()),
	})

	agent.PasswordHashed = ""
	agent.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"ID":      agent.ID,
		"name":    agent.Name,
		"email":   agent.Email,
		"agent":   agent,
		// "token":   token,
		"expires_in": int(middleware.TokenExpiry.Seconds()),
		"details":    agent.Details,
		"role":       agent.Role,
	})
}

func (h *Handler) VerifyAuth(c *gin.Context) {
	email, emailOk := c.Get("email")
	role, roleOk := c.Get("role")
	if !emailOk || !roleOk {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var agent models.Agent
	db := middleware.GetDB(c)
	if err := db.Where("email = ?", email.(string)).First(&agent); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	newToken, err := middleware.GenerateJWT(agent.ID, agent.Name, agent.Email, string(role.(string)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not refresh token"})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    newToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   int(middleware.TokenExpiry.Seconds()),
	})

	agent.PasswordHashed = ""
	agent.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"message":    "Authenticated",
		"agent":      agent,
		"token":      newToken,
		"role":       role,
		"expires_in": int(middleware.TokenExpiry.Seconds()),
	})
}

func (h *Handler) Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func (h *Handler) CheckEmailExists(c *gin.Context) {

	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	var agent models.Agent
	db := middleware.GetDB(c)
	if err := db.Where("email = ?", email).First(&agent); err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusOK, gin.H{"message": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": true})
}
func (h *Handler) CheckPhoneExists(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is required"})
		return
	}

	var agent models.Agent
	db := middleware.GetDB(c)

	if err := db.Where("details ->> 'phone' = ?", phone).First(&agent); err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusOK, gin.H{"message": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": true})
}

type EnableTOTPRequest struct {
	Email    string `json:"email" binding:"required"`
	TOTPCode string `json:"totp_code" binding:"required"`
}

func (h *Handler) EnableTOTP(c *gin.Context) {
	db := middleware.GetDB(c)

	var req EnableTOTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var agent models.Agent

	if err := db.Where("email = ?", req.Email).First(&agent); err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if !totp.Validate(req.TOTPCode, agent.Details.TOTPSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid TOTP code"})
		return
	}
	agent.Details.TOTPEnabled = true
	if err := db.Save(&agent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable 2FA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}
func (h *Handler) CheckMFA(c *gin.Context) {
	db := middleware.GetDB(c)
	email := c.GetHeader("user-email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required in the header"})
		return
	}

	var agent models.Agent
	if err := db.Where("email = ?", email).First(&agent); err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"totp_enabled": agent.Details.TOTPEnabled})
}
func (h *Handler) GetAgent(c *gin.Context) {
	db := middleware.GetDB(c)

	id := c.Param("id")
	fmt.Print("id", id)
	var agent models.Agent
	if err := db.Where("id=?", id).First(&agent); err != nil {
		return
	}
	agent.PasswordHashed = ""
	agent.Password = ""

	c.JSON(http.StatusOK, agent)
}

func (h *Handler) UpdateAgent(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Database not found in context",
			Code:  "DB_ERROR",
		})
		return
	}

	id := c.Param("id")
	var existingAgent models.Agent
	if err := db.Where("id = ?", id).First(&existingAgent); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Agent not found",
			Code:    "NOT_FOUND",
			Details: err.Error(),
		})
		return
	}

	var updateData models.Agent
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	if updateData.Email != existingAgent.Email {
		var emailCheck models.Agent
		if err := db.Where("email = ? AND id != ?", updateData.Email, id).First(&emailCheck).Error; err == nil {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Email already exists",
				Code:    "DUPLICATE_EMAIL",
				Details: "An agent with this email address already exists",
			})
			return
		}
	}

	existingAgent.Name = updateData.Name
	existingAgent.Email = updateData.Email
	existingAgent.Role = updateData.Role
	existingAgent.AssigneeType = updateData.AssigneeType
	existingAgent.Details = updateData.Details
	existingAgent.Superuser = updateData.Superuser

	if err := db.Save(&existingAgent); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update agent",
			Code:    "UPDATE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Agent updated successfully",
		Data:    existingAgent,
	})
}

func (h *Handler) UpdateSuperuserStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID",
			Code:    "INVALID_ID",
			Details: "ID cannot be empty",
		})
		return
	}

	db := middleware.GetDB(c)
	var existingAgent models.Agent
	if err := db.Where("id = ?", id).First(&existingAgent); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Agent not found",
			Code:    "NOT_FOUND",
			Details: "No agent with ID " + id,
		})
		return
	}

	var updateData struct {
		Superuser bool `json:"superuser"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	existingAgent.Superuser = updateData.Superuser

	if err := db.Save(&existingAgent); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update agent",
			Code:    "UPDATE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Superuser status updated successfully",
		Data:    existingAgent.Superuser,
	})
}

func (h *Handler) DeleteAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID",
			Code:    "INVALID_ID",
			Details: "Agent ID cannot be empty",
		})
		return
	}
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Database connection failed",
			Code:  "DB_ERROR",
		})
		return
	}

	var agent models.Agent
	if err := db.First(&agent, "id = ?", id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Agent not found",
			Code:  "NOT_FOUND",
		})
		return
	}

	if err := db.Delete(&agent); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete agent",
			Code:    "DELETE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Agent deleted successfully",
	})
}

func (h *Handler) ListAgents(c *gin.Context) {
	db := middleware.GetDB(c)

	page, pageSize := h.getPaginationParams(c)
	offset := (page - 1) * pageSize

	all := c.Query("all")

	excludeSelf := strings.ToLower(c.Query("exclude_self")) == "true" || c.Query("exclude_self") == "1"

	loggedInUserEmail := ""
	if excludeSelf {
		loggedInUserEmail = c.GetString("email")
		if loggedInUserEmail != "" {
			loggedInUserEmail = strings.ToLower(loggedInUserEmail)
		}
	}
	var agents []models.Agent
	var total int64
	countQuery := db.Model(&models.Agent{})
	if excludeSelf && loggedInUserEmail != "" {
		countQuery = countQuery.Where("LOWER(email) != ?", loggedInUserEmail)
	}
	if search := c.Query("search"); search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		countQuery = countQuery.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(role) LIKE ? OR LOWER(id) LIKE ?",
		searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if role := c.Query("role"); role != "" {
		countQuery = countQuery.Where("role = ?", role)
	}
	if assigneeType := c.Query("assignee_type"); assigneeType != "" {
		countQuery = countQuery.Where("assignee_type = ?", assigneeType)
	}
	if phone := c.Query("phone"); phone != "" {
		countQuery = countQuery.Where("details ->> 'phone' = ?", phone)
	}
	if err := countQuery.Count(&total).Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count agents"})
		return
	}
	if all == "true" {
		if err := countQuery.Find(&agents); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to fetch agents",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"data":        agents,
				"total":       total,
				"page":        nil,
				"page_size":   nil,
				"total_pages": nil,
			},
		})
		return
	}

	mainQuery := countQuery.
		Select("id, name, email, details, role, assignee_type, superuser,show_prices, details ->> 'phone' AS phone")

	if err := mainQuery.Offset(offset).Limit(pageSize).Find(&agents); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"data":        agents,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// func (h *Handler) UpdateShowPrices(c *gin.Context) {
// 	db := middleware.GetDB(c)
// 	if db == nil {
// 		c.JSON(http.StatusInternalServerError, ErrorResponse{
// 			Error: "Database not found — X-Tenant header may be missing or invalid",
// 			Code:  "DB_ERROR",
// 		})
// 		return
// 	}

// 	id := c.Param("id")
// 	if id == "" {
// 		c.JSON(http.StatusBadRequest, ErrorResponse{
// 			Error:   "Invalid ID",
// 			Code:    "INVALID_ID",
// 			Details: "Agent ID cannot be empty",
// 		})
// 		return
// 	}

// 	var body struct {
// 		ShowPrices bool `json:"show_prices"`
// 	}
// 	if err := c.ShouldBindJSON(&body); err != nil {
// 		c.JSON(http.StatusBadRequest, ErrorResponse{
// 			Error:   "Invalid request body",
// 			Code:    "VALIDATION_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	var agent models.Agent
// 	if err := db.Where("id = ?", id).First(&agent); err != nil {
// 		c.JSON(http.StatusNotFound, ErrorResponse{
// 			Error:   "Agent not found",
// 			Code:    "NOT_FOUND",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	agent.ShowPrices = body.ShowPrices
// 	if err := db.Save(&agent); err != nil {
// 		c.JSON(http.StatusInternalServerError, ErrorResponse{
// 			Error:   "Failed to update show_prices",
// 			Code:    "UPDATE_ERROR",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, SuccessResponse{
// 		Message: "Show prices updated successfully",
// 		Data:    gin.H{"show_prices": agent.ShowPrices},
// 	})
// }
