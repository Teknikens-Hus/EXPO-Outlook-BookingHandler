package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"net/smtp"

	cfghelper "github.com/Teknikens-Hus/EXPO-Outlook-BookingHandler/internal/conf"
	log "github.com/rs/zerolog/log"
)

var sentEmailsMutex sync.Mutex

type Overlap struct {
	resourceName    string
	expoBookingURL  string
	expoHumanNumber string
	expoEventName   string
	expoStartTime   time.Time
	expoEndTime     time.Time
	icsUID          string
	icsSummary      string
	icsStartTime    time.Time
	icsEndTime      time.Time
	icsName         string
}

type EventData struct {
	Summary string `json:"summary"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

type CalendarEvent struct {
	Summary    string
	Start      time.Time
	End        time.Time
	Reacurring bool
	UID        string
	TimeZone   string
}

func sendEmail(overlap Overlap, mailSettings cfghelper.MailSettings) error {
	foundRecipient := true
	toAddress, err := lookupEmail(overlap.icsSummary, &mailSettings)
	if err != nil {
		toAddress = mailSettings.FallbackEmail.Address
		log.Printf("Mail: Error looking up email: %s, sending to fallback: %s", err, toAddress)
		foundRecipient = false
	}
	const sentEmailsFile = "/app/data/sent_emails.txt"
	sent, err := hasEmailBeenSent(overlap.icsUID, sentEmailsFile)
	if err != nil {
		log.Printf("Mail: Error checking if email has been sent: %v", err)
	}
	if sent {
		log.Printf("Mail: Email for %s already sent, skipping", overlap.icsUID)
		return nil
	}
	var subject string
	var htmlContent string
	if foundRecipient {
		subject = mailSettings.Subject
		htmlContent, err = formatContentHTML(mailSettings.MailContent, overlap)
		if err != nil {
			log.Printf("Mail: Error formatting fallback content: %v", err)
			return err
		}
	} else {
		// Use fallback
		subject = mailSettings.Subject + "-Fallback"
		htmlContent, err = formatContentHTML(mailSettings.MailContentFallback, overlap)
		if err != nil {
			log.Printf("Mail: Error formatting fallback content: %v", err)
			return err
		}
	}
	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"
	message := "From: " + mailSettings.From.Address + "\r\n" +
		"To: " + toAddress + "\r\n" +
		"Subject: " + subject + "\r\n" +
		headers + "\r\n" +
		"\r\n" +
		htmlContent
	if !mailSettings.SendEmails {
		log.Print("Mail: Not sending email, SendEmails is set to false")
		log.Printf("Mail: Would have sent email to: %s with subject: %s", toAddress, subject)
		markEmailAsSent(overlap.icsUID, sentEmailsFile)
		return nil
	}
	SMTP_Password := os.Getenv("SMTP_PASSWORD")
	if SMTP_Password == "" {
		return errors.New("SMTP Password is not set")
	}
	SMTP_USERNAME := os.Getenv("SMTP_USERNAME")
	if SMTP_USERNAME == "" {
		return errors.New("SMTP Username is not set")
	}
	SMTP_HOST := os.Getenv("SMTP_HOST")
	if SMTP_HOST == "" {
		return errors.New("SMTP Host is not set")
	}
	SMTP_PORT := os.Getenv("SMTP_PORT") // Default to 587
	if SMTP_PORT == "" {
		SMTP_PORT = "587"
	}
	auth := smtp.PlainAuth(
		"",
		SMTP_USERNAME,
		SMTP_Password,
		SMTP_HOST)
	log.Printf("Mail: Sending email to: %s from: %s using: %s", toAddress, mailSettings.From.Address, SMTP_HOST)
	err = smtp.SendMail(fmt.Sprintf("%s:%s", SMTP_HOST, SMTP_PORT), auth, mailSettings.From.Address, []string{toAddress}, []byte(message))
	if err != nil {
		log.Printf("Mail: Error sending email: %v", err)
		return err
	} else {
		log.Printf("Mail: Email sent to: %s, from %s", toAddress, mailSettings.From.Address)
		markEmailAsSent(overlap.icsUID, sentEmailsFile)
	}

	return nil
}

func RegisterOverlap(newOverlap Overlap, mailSettings cfghelper.MailSettings) {
	log.Printf(("Got new overlap for EXPO Booking %s in Calendar %s with summary: %s"), newOverlap.expoHumanNumber, newOverlap.icsName, newOverlap.icsSummary)
	err := sendEmail(newOverlap, mailSettings)
	if err != nil {
		log.Printf("Error sending email for overlap: %v", err)
	}
}

func lookupEmail(icsSummary string, mailSettings *cfghelper.MailSettings) (string, error) {
	icsSummary = strings.ToLower(strings.ReplaceAll(icsSummary, " ", ""))
	log.Print("Mail: Looking up email for summary: ", icsSummary)
	for _, mapping := range mailSettings.Mappings {
		if strings.ToLower(strings.ReplaceAll(mapping.IcsSummary, " ", "")) == icsSummary {
			return mapping.Address, nil
		} else {
			log.Print("Mail: No match for summary: ", mapping.IcsSummary)
		}
	}
	return "", fmt.Errorf("no email found for summary: %s", icsSummary)
}

func hasEmailBeenSent(icsUID string, filename string) (bool, error) {
	sentEmailsMutex.Lock()
	defer sentEmailsMutex.Unlock()
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return false, err
	}
	defer file.Close()
	var line string
	for {
		_, err := fmt.Fscanf(file, "%s\n", &line)
		if err != nil {
			if err == io.EOF {
				return false, nil
			}
			return true, fmt.Errorf("error reading file %s: %w", filename, err) // if we have error, lets be safe and not send emails over and over
		}
		if line == icsUID {
			return true, nil
		}
	}
}

func markEmailAsSent(icsUID string, filename string) {
	sentEmailsMutex.Lock()
	defer sentEmailsMutex.Unlock()
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	fmt.Fprintln(file, icsUID)
}

func formatContentHTML(contentTemplate string, overlap Overlap) (string, error) {
	template, err := template.New("email").Parse(contentTemplate)
	if err != nil {
		return fmt.Sprintf("Error parsing content template: %v", err), err
	}
	data := map[string]interface{}{
		"Summary":     overlap.icsSummary,
		"Resource":    overlap.resourceName,
		"Start":       overlap.icsStartTime.Format(time.RFC3339),
		"End":         overlap.icsEndTime.Format(time.RFC3339),
		"BookingURL":  overlap.expoBookingURL,
		"HumanNumber": overlap.expoHumanNumber,
	}
	var buf bytes.Buffer
	if err := template.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Error executing content template: %v", err), err
	}
	return buf.String(), nil
}
