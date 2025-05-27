package main

import (
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // without force loading of timezone data the TZ environment variable is not applied correctly

	log "github.com/rs/zerolog/log"
)

func main() {
	log.Print("EXPO Outlook BookingHandler starting...")
	// Manually update timezone from TZ env variable
	if tz := os.Getenv("TZ"); tz != "" {
		var err error
		time.Local, err = time.LoadLocation(tz)
		if err != nil {
			log.Printf("error loading location '%s': %v\n", tz, err)
		}
	} else {
		log.Print("TZ environment variable not found")
	}
	log.Printf("Timezone set to: %s\n", time.Local)
	log.Printf("Current time: %s\n", time.Now().Format(time.RFC3339))

	// Setup EXPO
	expoConfig, err := SetupEXPO()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to setup EXPO")
	}
	// Get settings from config file
	settings, err := LoadConfigFile()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get settings")
	}
	// Check if we have any calendars to check
	if settings.ICSConfig.Calendars == nil {
		log.Print("ICS: No ICS configurations found in the config file")
		return
	}

	checkOverlaps(expoConfig, settings)
	setupTicker(20, expoConfig, settings)
	// Keep the application running
	select {}
}

func checkOverlaps(expoConfig *EXPOConfig, settings *ConfigSettings) {
	// Get today -1 day and the last day of the month
	start, end := GetMonthDateRange()
	// Fetch bookings from EXPO
	expoBookings := GetNewBookings(expoConfig, start, end)
	expoBookings = filterConfirmedBookings(expoBookings)
	expoBookings = filterBookingWithResource(expoBookings, settings.Resources)
	bookingsURLSuffix := "/administration/bookings/"
	_, err := url.Parse(expoConfig.EXPOURL + bookingsURLSuffix)
	if err != nil {
		log.Print("Error parsing EXPO URL: ", err)
		return
	}
	// Loop through the calendars and get the events
	for i, ics := range settings.ICSConfig.Calendars {
		log.Print("Fetching calendar: ", ics.Name)
		events, err := GetCalendarEventsFromICS(&settings.ICSConfig.Calendars[i], start, end)
		if err != nil {
			log.Print("ICS: Error getting calendar events: ", err)
		}
		log.Print("ICS: Found ", len(events), " events in calendar: ", ics.Name)
		// Loop through the events and check for overlaps
		for _, event := range events {
			//log.Print("ICS: Event: ", event.Summary, " Start: ", event.Start.Format(time.RFC3339), " End: ", event.End.Format(time.RFC3339))
			if event.Reacurring {
				log.Print("ICS: Event is recurring")
			}
			// Loop through all bookings and check for overlaps with the current event
			for _, booking := range expoBookings {
				for _, resourceMap := range settings.Resources {
					if strings.EqualFold(ics.Name, resourceMap.EXPOResourceName) {
						//log.Printf("Event %s in calendar %s matches resourceMap %s", event.Summary, ics.Name, resourceMap.EXPOResourceName)
						doesOverlap, overlapEventName, eventStartTime, eventEndTime := doesBookingResourceOverlap(booking, event.Start, event.End, ics.Name)
						if doesOverlap {
							bookingURL := expoConfig.EXPOURL + bookingsURLSuffix + strconv.Itoa(booking.ID)
							RegisterOverlap(Overlap{
								resourceMap.Name,
								bookingURL,
								booking.HumanNumber,
								overlapEventName,
								eventStartTime,
								eventEndTime,
								event.UID,
								event.Summary,
								event.Start,
								event.End,
								ics.Name,
							}, settings.MailSettings)
							break
						}
					}
				}
			}

		}
	}
}

func GetMonthDateRange() (time.Time, time.Time) {
	// Calculate the first and last day of the current month
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastDay := firstDay.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	start := now.Add(-1 * time.Hour * 24)
	end := lastDay
	log.Print("Start date: ", start)
	log.Print("End date: ", end)
	return start, end
}

func setupTicker(interval int, expoConfig *EXPOConfig, settings *ConfigSettings) {
	log.Print("Setting up ticker with interval ", interval, " seconds")
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Print(("Ticker triggered, checking overlaps..."))
				checkOverlaps(expoConfig, settings)
			}
		}
	}()
}
