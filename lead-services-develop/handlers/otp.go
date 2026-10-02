package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/store"
	"github.com/tuliporg/agent-management-api/utils"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type otpReq struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type otpResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// otpEmailBody renders the HTML email sent to the user for a password reset.
func otpEmailBody(code string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="margin:0;padding:0;background-color:#f4f5f7;font-family:'Helvetica Neue',Arial,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="padding:40px 0;">
    <tr>
      <td align="center">
        <table width="480" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 1px 3px rgba(0,0,0,0.1);">
          <tr>
            <td style="padding:32px;text-align:center;">
              <p style="font-size:18px;font-weight:600;color:#111827;margin:0 0 20px;">Password reset</p>
              <p style="font-size:14px;color:#4b5563;margin:0 0 20px;line-height:1.6;">Use the following one-time code to reset your password.</p>
              <div style="background:#f3f4f6;border-radius:6px;padding:18px;margin-bottom:16px;">
                <span style="font-size:30px;letter-spacing:8px;font-weight:700;color:#111827;">%s</span>
              </div>
              <p style="font-size:13px;color:#4b5563;margin:0 0 20px;">Valid for 5 minutes.</p>
              <div style="border-top:1px solid #e5e7eb;padding-top:16px;text-align:left;">
                <p style="font-size:12px;color:#9ca3af;margin:0 0 8px;line-height:1.6;">If you didn't request a password reset, please ignore this email.</p>
                <p style="font-size:12px;color:#9ca3af;margin:0;">Never share this code with anyone.</p>
              </div>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, code)
}

// SendOTP sends an OTP to the given email
func (h *Handler) SendOTP(c *gin.Context) {

	var req otpReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Code:    "VALIDATION_ERROR",
			Details: "Email is required",
		})
		return
	}

	database := middleware.GetDB(c)

	// Generate OTP
	code := fmt.Sprintf("%04d", rand.Intn(10000))

	// Save OTP in store
	if err := store.Save(database, req.Email, code); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to save OTP",
			Code:    "STORE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Send Email
	body := otpEmailBody(code)
	tenant := middleware.GetTenant(c)

	if err := utils.SendEmail(tenant, req.Email, "Password reset code", body); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to send OTP email",
			Code:    "EMAIL_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, otpResp{
		Success: true,
		Message: "OTP sent successfully",
	})
}

// VerifyOTP verifies the OTP for the given email
func (h *Handler) VerifyOTP(c *gin.Context) {
	var req otpReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Code == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Code:    "VALIDATION_ERROR",
			Details: "Email and code are required",
		})
		return
	}

	database := middleware.GetDB(c)

	// Verify OTP
	valid, err := store.Verify(database, req.Email, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to verify OTP",
			Code:    "STORE_ERROR",
			Details: err.Error(),
		})
		return
	}

	msg := "Invalid or expired OTP"
	if valid {
		msg = "OTP verified successfully"
	}

	c.JSON(http.StatusOK, otpResp{
		Success: valid,
		Message: msg,
	})
}