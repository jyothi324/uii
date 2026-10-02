package utils
import "github.com/pquerna/otp/totp"

func ValidateTOTP(code, secret string) bool {
    return totp.Validate(code, secret)
}