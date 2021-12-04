package pkg

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SendEmail sends an email
func SendEmail(to string) (code string, err error) {
	var client *smtp.Client
	subject := os.Getenv("EMAIL_SUBJECT")
	from := os.Getenv("EMAIL_FROM")
	smtpServer := os.Getenv("SMTP_SERVER")
	codeLength := os.Getenv("VERIFY_CODE_LENGTH")
	if gin.Mode() == gin.ReleaseMode {
		if smtpServer == "" {
			err = fmt.Errorf("SMTP_SERVER environment variable is not set")
		}
		if subject == "" {
			err = fmt.Errorf("SMTP_SUBJECT environment variable is not set")
		}
		if from == "" {
			err = fmt.Errorf("SMTP_FROM environment variable is not set")
		}
		if codeLength == "" {
			err = fmt.Errorf("VERIFY_CODE_LENGTH environment variable is not set")
		}
		if err != nil {
			return
		}
	} else {
		codeLength = "6"
	}

	length, err := strconv.Atoi(codeLength)
	if err != nil {
		return
	}

	code = randString(length)
	if gin.Mode() != gin.ReleaseMode {
		return
	}

	client, err = smtp.Dial(smtpServer)
	if err != nil {
		return
	}
	defer client.Close()

	client.Mail(from)
	client.Rcpt(to)

	wc, err := client.Data()
	if err != nil {
		return
	}
	defer wc.Close()

	buf := bytes.NewBufferString("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		"Verify code is " + code + "r\n")

	if _, err = buf.WriteTo(wc); err != nil {
		return
	}

	return
}
