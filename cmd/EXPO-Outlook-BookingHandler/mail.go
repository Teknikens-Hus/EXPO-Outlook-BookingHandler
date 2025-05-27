package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

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

func sendEmail(overlap Overlap, mailSettings MailSettings) error {
	foundRecipient := true
	email, err := lookupEmail(overlap.icsSummary, &mailSettings)
	if err != nil {
		email = mailSettings.FallbackEmail.Address
		log.Printf("Mail: Error looking up email: %s, sending to fallback: %s", err, email)
		foundRecipient = false
	}
	log.Printf("Mail: Sending email to: %s from: %s", email, mailSettings.From.Address)
	from := mail.NewEmail(mailSettings.From.Name, mailSettings.From.Address)
	var subject string
	var to *mail.Email
	var htmlContent string
	if foundRecipient {
		subject = mailSettings.Subject
		to = mail.NewEmail(overlap.icsSummary, email)
		htmlContent = fmt.Sprintf("Hej!<br> %s<br>Din bokning av %s<br>%s till %s<br>Överlappar med EXPO bokning <a href=%s target='_blank'>%s</a><br>Överväg en annan lokal eller tid.", overlap.icsSummary, overlap.resourceName, overlap.icsStartTime, overlap.icsEndTime, overlap.expoBookingURL, overlap.expoHumanNumber)
	} else {
		// Use fallback
		subject = mailSettings.Subject + " - Fallback"
		to = mail.NewEmail(mailSettings.FallbackEmail.Name, mailSettings.FallbackEmail.Address)
		htmlContent = fmt.Sprintf("Hittade inte mail för %s<br> Bokning av %s<br>%s till %s<br>Överlappar med EXPO bokning <a href=%s target='_blank'>%s</a><br>Överväg lägga till personen i Summary->email mappningen för EXPO-Outlook-BookingHandler.", overlap.icsSummary, overlap.resourceName, overlap.icsStartTime, overlap.icsEndTime, overlap.expoBookingURL, overlap.expoHumanNumber)
	}
	plainTextContent := ""
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	if !mailSettings.SendEmails {
		log.Print("Mail: Not sending email, SendEmails is set to false")
		log.Printf("Mail: Would have sent email to: %s with subject: %s", email, subject)
		return nil
	}
	APIkey := os.Getenv("SENDGRID_APIKEY")
	if APIkey == "" {
		return errors.New("SendGrid APIkey is not set")
	}
	client := sendgrid.NewSendClient(APIkey)
	response, err := client.Send(message)
	if err != nil {
		log.Print("Mail: Error sending email: ", err)
		return err
	} else {
		log.Printf("Mail: Email sent to: %s with status code: %d, from %s", email, response.StatusCode, mailSettings.From.Address)
	}
	return nil
}

func RegisterOverlap(newOverlap Overlap, mailSettings MailSettings) {
	log.Printf(("Got new overlap for EXPO Booking %s in Calendar %s with summary: %s"), newOverlap.expoHumanNumber, newOverlap.icsName, newOverlap.icsSummary)
	sendEmail(newOverlap, mailSettings)
}

func lookupEmail(icsSummary string, mailSettings *MailSettings) (string, error) {
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
