package utils

import (
	"fmt"
	"net/smtp"
	"os"
	"bytes"
	"encoding/base64"

	//"github.com/joho/godotenv"
)
func getSMTPCredentials(tenant string) (string, string) {
	if tenant == "acergranite" {
		return os.Getenv("ACER_SMTP_USER"), os.Getenv("ACER_SMTP_PASS")
	}
	return os.Getenv("DEFAULT_SMTP_USER"), os.Getenv("DEFAULT_SMTP_PASS")
}

func SendEmail(tenant,to, subject, body string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	user, pass := getSMTPCredentials(tenant)
	auth := smtp.PlainAuth("", user, pass, smtpHost)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s\r\n",
		user, to, subject, body,
	))
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	return smtp.SendMail(addr, auth, user, []string{to}, msg)
}

//
// for sending report with attachment
//
func SendEmailWithAttachment(tenant, to, subject, body, filename string, fileBytes []byte) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	user, pass := getSMTPCredentials(tenant)

	auth := smtp.PlainAuth("", user, pass, smtpHost)

	boundary := "BOUNDARY-REPORT-EMAIL"

	var msg bytes.Buffer

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", user))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	msg.WriteString("\r\n")

	// Body part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
	msg.WriteString(body + "\r\n")

	// Attachment part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: application/octet-stream\r\n")
	msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
	msg.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(fileBytes)))
	base64.StdEncoding.Encode(encoded, fileBytes)

	// split base64 into lines of 76 chars (SMTP rule)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		msg.Write(encoded[i:end])
		msg.WriteString("\r\n")
	}

	msg.WriteString(fmt.Sprintf("--%s--", boundary))

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	return smtp.SendMail(addr, auth, user, []string{to}, msg.Bytes())
}

// EmailAttachment represents a single file to attach to an outgoing email.
type EmailAttachment struct {
	Name  string
	Bytes []byte
}

//
// for sending report with multiple attachments in a single email
//
func SendEmailWithAttachments(tenant,to, subject, body string, attachments []EmailAttachment) error {
	if len(attachments) == 0 {
		return SendEmail(tenant,to, subject, body)
	}
	if len(attachments) == 1 {
		return SendEmailWithAttachment(tenant, to, subject, body, attachments[0].Name, attachments[0].Bytes)
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	user, pass := getSMTPCredentials(tenant)

	auth := smtp.PlainAuth("", user, pass, smtpHost)

	boundary := "BOUNDARY-REPORT-EMAIL"

	var msg bytes.Buffer

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", user))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	msg.WriteString("\r\n")

	// Body part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
	msg.WriteString(body + "\r\n")

	// One attachment part per file
	for _, att := range attachments {
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: application/octet-stream\r\n")
		msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", att.Name))
		msg.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

		encoded := make([]byte, base64.StdEncoding.EncodedLen(len(att.Bytes)))
		base64.StdEncoding.Encode(encoded, att.Bytes)

		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			msg.Write(encoded[i:end])
			msg.WriteString("\r\n")
		}
	}

	msg.WriteString(fmt.Sprintf("--%s--", boundary))

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	return smtp.SendMail(addr, auth, user, []string{to}, msg.Bytes())
}