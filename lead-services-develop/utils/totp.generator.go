package utils

import "github.com/pquerna/otp/totp"

func GenerateTOTP(email string, enabled bool) (string, string, error) {
	if !enabled {
		return "", "", nil
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Tulip",
		AccountName: email,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}
