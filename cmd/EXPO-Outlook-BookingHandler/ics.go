package main

import (
	"fmt"
	"net/http"
	"time"

	cfghelper "github.com/Teknikens-Hus/EXPO-Outlook-BookingHandler/internal/conf"
	"github.com/apognu/gocal"
)

func GetCalendarEventsFromICS(calConfig *cfghelper.CalendarConfig, start, end time.Time) ([]CalendarEvent, error) {
	resp, err := http.Get(calConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w for calendar: %s", err, calConfig.Name)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	// Check if the content type is text/calendar, if its text/html the URL is probably wrong or expired
	if contentType := resp.Header.Get("Content-Type"); contentType != "text/calendar; charset=utf-8" {
		return nil, fmt.Errorf("unexpected content type: %s", contentType)
	}
	calendar := gocal.NewParser(resp.Body)

	// Here we can map the timezone IDs from the ICS file to the Go time.Location
	// This is useful if yourtimezone cant be resolved by Go
	var tzMapping = map[string]string{
		"W. Europe Standard Time": "Europe/Stockholm",
	}
	gocal.SetTZMapper(func(s string) (*time.Location, error) {
		if tzid, ok := tzMapping[s]; ok {
			return time.LoadLocation(tzid)
		}
		return nil, fmt.Errorf("unknown timezone: %s", s)
	})
	// Set the start and end date for the calendar parser (Which event dates to parse)
	calendar.Start, calendar.End = &start, &end
	err = calendar.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse calendar: %w", err)
	}
	// Convert the gocal events to our own CalendarEvent struct
	var events []CalendarEvent
	for _, event := range calendar.Events {
		events = append(events, CalendarEvent{
			Summary:    event.Summary,
			Start:      *event.Start,
			End:        *event.End,
			Reacurring: event.IsRecurring,
			UID:        event.Uid,
		})

	}
	return events, nil
}
