package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strings"

	"github.com/ritikarora108/ai-powered-sast-tool/backend/db"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"go.uber.org/zap"
)

// EmailService handles sending email notifications
type EmailService struct {
	smtpServer   string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	dbQueries    *db.Queries
}

// NewEmailService creates a new instance of EmailService
func NewEmailService(dbQueries *db.Queries) *EmailService {
	return &EmailService{
		smtpServer:   os.Getenv("SMTP_SERVER"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    os.Getenv("FROM_EMAIL"),
		dbQueries:    dbQueries,
	}
}

// ScanCompletionEmailData contains data needed for the scan completion email template
type ScanCompletionEmailData struct {
	RepositoryName string
	DashboardURL   string
	VulnCount      int
}

// SendScanCompletionEmail sends a notification email that a repository scan is complete
func (s *EmailService) SendScanCompletionEmail(userEmail, repositoryName, repositoryID string, vulnCount int) error {
	log := logger.Get()

	if s.smtpServer == "" || s.smtpPort == "" || s.smtpUsername == "" ||
		s.smtpPassword == "" || s.fromEmail == "" {
		return fmt.Errorf("email service is not properly configured")
	}

	// Create email data
	dashboardURL := os.Getenv("DASHBOARD_URL")
	if dashboardURL == "" {
		dashboardURL = "http://localhost:3000"
	}

	repoDetailsURL := fmt.Sprintf("%s/dashboard/repos/%s", dashboardURL, repositoryID)

	data := ScanCompletionEmailData{
		RepositoryName: repositoryName,
		DashboardURL:   repoDetailsURL,
		VulnCount:      vulnCount,
	}

	// Parse email template
	emailTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Scan Results Available</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .container {
            background-color: #ffffff;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
            padding: 30px;
        }
        .header {
            text-align: center;
            margin-bottom: 20px;
        }
        .logo {
            max-width: 120px;
            margin-bottom: 15px;
        }
        h1 {
            color: #2563eb;
            font-size: 24px;
            margin-bottom: 15px;
        }
        .content {
            margin-bottom: 25px;
        }
        .button {
            display: inline-block;
            background-color: #2563eb;
            color: white;
            text-decoration: none;
            padding: 12px 25px;
            border-radius: 6px;
            font-weight: 600;
            margin: 15px 0;
        }
        .button:hover {
            background-color: #1d4ed8;
        }
        .footer {
            margin-top: 30px;
            text-align: center;
            font-size: 14px;
            color: #6b7280;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Security Scan Results Available</h1>
        </div>
        <div class="content">
            <p>Hello,</p>
            <p>We've completed the security scan for your repository <strong>{{.RepositoryName}}</strong>.</p>
            <p>{{if gt .VulnCount 0}}
                We found <strong>{{.VulnCount}} potential security issues</strong> that you should review.
            {{else}}
                Good news! No security issues were found in your repository.
            {{end}}</p>
            <p>View the detailed results on your dashboard:</p>
            <p style="text-align: center;">
                <a href="{{.DashboardURL}}" class="button">View Scan Results</a>
            </p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>
`

	// Execute template with data
	var body bytes.Buffer
	tmpl, err := template.New("scanEmail").Parse(emailTemplate)
	if err != nil {
		log.Error("Failed to parse email template", zap.Error(err))
		return err
	}

	if err := tmpl.Execute(&body, data); err != nil {
		log.Error("Failed to execute email template", zap.Error(err))
		return err
	}

	// Compose email
	to := []string{userEmail}
	subject := fmt.Sprintf("Security Scan Results Available - %s", repositoryName)

	headers := make(map[string]string)
	headers["From"] = s.fromEmail
	headers["To"] = userEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var message bytes.Buffer
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.Write(body.Bytes())

	// Connect to SMTP server and send email
	addr := fmt.Sprintf("%s:%s", s.smtpServer, s.smtpPort)
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpServer)

	err = smtp.SendMail(addr, auth, s.fromEmail, to, message.Bytes())
	if err != nil {
		log.Error("Failed to send email",
			zap.String("to", userEmail),
			zap.String("subject", subject),
			zap.Error(err))
		return err
	}

	log.Info("Scan completion email sent successfully",
		zap.String("to", userEmail),
		zap.String("repository", repositoryName))

	return nil
}

// SendBulkScanCompletionEmail sends a notification email to multiple recipients
func (s *EmailService) SendBulkScanCompletionEmail(userEmails []string, repositoryName, repositoryID string, vulnCount int) error {
	log := logger.Get()

	if len(userEmails) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	if s.smtpServer == "" || s.smtpPort == "" || s.smtpUsername == "" ||
		s.smtpPassword == "" || s.fromEmail == "" {
		return fmt.Errorf("email service is not properly configured")
	}

	// Create email data
	dashboardURL := os.Getenv("DASHBOARD_URL")
	if dashboardURL == "" {
		dashboardURL = "http://localhost:3000"
	}

	repoDetailsURL := fmt.Sprintf("%s/dashboard/repos/%s", dashboardURL, repositoryID)

	data := ScanCompletionEmailData{
		RepositoryName: repositoryName,
		DashboardURL:   repoDetailsURL,
		VulnCount:      vulnCount,
	}

	// Parse email template - same as above
	emailTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Scan Results Available</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .container {
            background-color: #ffffff;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
            padding: 30px;
        }
        .header {
            text-align: center;
            margin-bottom: 20px;
        }
        .logo {
            max-width: 120px;
            margin-bottom: 15px;
        }
        h1 {
            color: #2563eb;
            font-size: 24px;
            margin-bottom: 15px;
        }
        .content {
            margin-bottom: 25px;
        }
        .button {
            display: inline-block;
            background-color: #2563eb;
            color: white;
            text-decoration: none;
            padding: 12px 25px;
            border-radius: 6px;
            font-weight: 600;
            margin: 15px 0;
        }
        .button:hover {
            background-color: #1d4ed8;
        }
        .footer {
            margin-top: 30px;
            text-align: center;
            font-size: 14px;
            color: #6b7280;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Security Scan Results Available</h1>
        </div>
        <div class="content">
            <p>Hello,</p>
            <p>We've completed the security scan for repository <strong>{{.RepositoryName}}</strong>.</p>
            <p>{{if gt .VulnCount 0}}
                We found <strong>{{.VulnCount}} potential security issues</strong> that should be reviewed.
            {{else}}
                Good news! No security issues were found in this repository.
            {{end}}</p>
            <p>View the detailed results on the dashboard:</p>
            <p style="text-align: center;">
                <a href="{{.DashboardURL}}" class="button">View Scan Results</a>
            </p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>
`

	// Execute template with data
	var body bytes.Buffer
	tmpl, err := template.New("scanEmail").Parse(emailTemplate)
	if err != nil {
		log.Error("Failed to parse email template", zap.Error(err))
		return err
	}

	if err := tmpl.Execute(&body, data); err != nil {
		log.Error("Failed to execute email template", zap.Error(err))
		return err
	}

	// Compose email with BCC for multiple recipients
	subject := fmt.Sprintf("Security Scan Results Available - %s", repositoryName)

	headers := make(map[string]string)
	headers["From"] = s.fromEmail
	headers["To"] = s.fromEmail                     // Set the main recipient as the from email
	headers["Bcc"] = strings.Join(userEmails, ", ") // Add all recipients as BCC
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var message bytes.Buffer
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.Write(body.Bytes())

	// Connect to SMTP server and send email
	addr := fmt.Sprintf("%s:%s", s.smtpServer, s.smtpPort)
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpServer)

	// For BCC, we need to include the from address as the recipient in the SMTP call
	// but the actual recipients will be those in the BCC header
	recipientList := append([]string{s.fromEmail}, userEmails...)
	err = smtp.SendMail(addr, auth, s.fromEmail, recipientList, message.Bytes())
	if err != nil {
		log.Error("Failed to send bulk email",
			zap.Strings("to", userEmails),
			zap.String("subject", subject),
			zap.Error(err))
		return err
	}

	log.Info("Scan completion email sent successfully to multiple recipients",
		zap.Strings("to", userEmails),
		zap.String("repository", repositoryName))

	return nil
}
