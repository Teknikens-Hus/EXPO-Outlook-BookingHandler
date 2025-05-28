package main

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/machinebox/graphql"
	log "github.com/rs/zerolog/log"
)

type EXPOConfig struct {
	EXPOURL   string
	EXPOToken string
	QUERY     string
}

func SetupEXPO() (*EXPOConfig, error) {
	var expoURL string = os.Getenv("EXPO_URL")
	if expoURL == "" {
		return nil, errors.New("EXPO_URL env variable not set or empty")
	}
	log.Print("EXPO_URL: ", expoURL)

	var expoToken string = os.Getenv("EXPO_TOKEN")
	if expoToken == "" {
		return nil, errors.New("EXPO_TOKEN env variable not set or empty")
	}
	log.Print("EXPO_TOKEN: ", expoToken)

	// Load the query from a file
	queryFile, err := os.ReadFile("query-booking.graphql")
	if err != nil {
		return nil, errors.New("failed to read query file")
	}
	query := string(queryFile)
	return &EXPOConfig{expoURL, expoToken, query}, nil
}

func GetNewBookings(config *EXPOConfig, startTime time.Time, endTime time.Time) []QueryUserResponseBookingNode {
	expoBookings, err := fetchEXPOBooking(config.EXPOURL, config.QUERY, startTime, endTime, config.EXPOToken)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to fetch EXPO bookings")
	}
	log.Print("EXPO bookings fetched successfully")
	return expoBookings
}

func fetchEXPOBooking(expoURL string, query string, startDate time.Time, endDate time.Time, expoToken string) ([]QueryUserResponseBookingNode, error) {
	var allNodes []QueryUserResponseBookingNode
	var cursor *string
	apiEndpoint := "/api/v3/graphql"
	_, err := url.Parse(expoURL + apiEndpoint)
	if err != nil {
		log.Print("Error parsing EXPO URL: ", err)
		return nil, err
	}
	for { // Infinite loop
		client := graphql.NewClient(expoURL + apiEndpoint)
		request := graphql.NewRequest(query)
		request.Header.Set("Authorization", "Bearer "+expoToken)
		request.Header.Set("Content-Type", "application/json")
		request.Var("startAtGteq", startDate.Format(time.RFC3339))
		request.Var("endAtLteq", endDate.Format(time.RFC3339))
		if cursor != nil {
			request.Var("cursor", *cursor)
		}

		var response QueryUserResponse
		err := client.Run(context.Background(), request, &response) // TODO rewrite in standard http client to handle unauthorized errors better
		if err != nil {
			log.Printf("Fetched, but error occurred: %v", err)
			return nil, err
		}

		allNodes = append(allNodes, response.Bookings.Nodes...)

		if !response.Bookings.PageInfo.HasNextPage {
			log.Printf("Fetched %d bookings, total: %d", len(response.Bookings.Nodes), response.Bookings.TotalNodeCount)
			break
		}
		cursor = &response.Bookings.PageInfo.EndCursor
		log.Printf("Fetched %d bookings, total: %d, next cursor: %s", len(response.Bookings.Nodes), response.Bookings.TotalNodeCount, *cursor)
	}

	return allNodes, nil
}

func filterConfirmedBookings(bookings []QueryUserResponseBookingNode) []QueryUserResponseBookingNode {
	var filteredBookings []QueryUserResponseBookingNode
	for _, booking := range bookings {
		if booking.State == "confirmed" {
			filteredBookings = append(filteredBookings, booking)
		}
	}
	log.Printf("Filtered confirmed bookings: %d, removed %d bookings", len(filteredBookings), len(bookings)-len(filteredBookings))
	return filteredBookings
}

func filterBookingWithResource(bookings []QueryUserResponseBookingNode, monitoredResourceNames []string) []QueryUserResponseBookingNode {
	var filteredBookings []QueryUserResponseBookingNode
	seen := make(map[string]bool)
	if len(monitoredResourceNames) == 0 {
		log.Print("No monitored resource names found, returning all bookings")
		return bookings
	}
	for _, booking := range bookings {
		for _, reservation := range booking.Reservations.Nodes {
			if reservation.Reservationable != nil {
				if reservation.Reservationable.Event.EventAllocation.EventAllocationResources.TotalNodeCount > 0 {
					for _, resource := range reservation.Reservationable.Event.EventAllocation.EventAllocationResources.Nodes {
						for _, monResource := range monitoredResourceNames {
							if strings.EqualFold(resource.Resource.Name, monResource) {
								if !seen[booking.HumanNumber] {
									filteredBookings = append(filteredBookings, booking)
									seen[booking.HumanNumber] = true
								}
								break
							}
						}
					}
				}
			} else {
				log.Printf("Booking: %s, has no reservationable event", booking.HumanNumber)
			}
		}
	}
	log.Printf("Filtered bookings with resources: %d, removed %d bookings", len(filteredBookings), len(bookings)-len(filteredBookings))
	return filteredBookings
}

func doesBookingResourceOverlap(booking QueryUserResponseBookingNode, startDate time.Time, endDate time.Time, resourceName string) (bool, string, time.Time, time.Time) {
	for _, reservation := range booking.Reservations.Nodes {
		if reservation.Reservationable != nil {
			if reservation.Reservationable.Event.StartAt.Before(endDate) && reservation.Reservationable.Event.EndAt.After(startDate) {
				for _, resource := range reservation.Reservationable.Event.EventAllocation.EventAllocationResources.Nodes {
					if resource.Resource.Name == resourceName {
						log.Printf("Found overlap for booking: %s with booking resource: %s, Event name: %s was looking for resource: %s", booking.HumanNumber, resource.Resource.Name, reservation.Reservationable.Event.Name, resourceName)
						return true, reservation.Reservationable.Event.Name, reservation.Reservationable.Event.StartAt, reservation.Reservationable.Event.EndAt
					}
				}
			}
		}
	}
	return false, "Not found", time.Time{}, time.Time{}
}

type QueryUserResponseBookingNode = struct {
	HumanNumber string
	ID          int
	State       string
	Email       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	BookingType struct {
		Name string
	}
	Booker struct {
		Customer struct {
			Name         string
			CustomerType struct {
				Name string
			}
		}
	}
	Reservations struct {
		Nodes []struct {
			Offer struct {
				Name string
			}
			Reservationable *struct {
				Event struct {
					Name            string
					StartAt         time.Time
					EndAt           time.Time
					EventAllocation struct {
						EventAllocationResources struct {
							TotalNodeCount int
							Nodes          []struct {
								Resource struct {
									Name         string
									ResourceType struct {
										Name string
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

type QueryUserResponse struct {
	Bookings struct {
		TotalNodeCount int
		TotalPageCount int
		PageInfo       struct {
			HasNextPage     bool
			EndCursor       string
			StartCursor     string
			HasPreviousPage bool
		}
		Nodes []struct {
			HumanNumber string
			ID          int
			State       string
			Email       string
			CreatedAt   time.Time
			UpdatedAt   time.Time
			BookingType struct {
				Name string
			}
			Booker struct {
				Customer struct {
					Name         string
					CustomerType struct {
						Name string
					}
				}
			}
			Reservations struct {
				Nodes []struct {
					Offer struct {
						Name string
					}
					Reservationable *struct {
						Event struct {
							Name            string
							StartAt         time.Time
							EndAt           time.Time
							EventAllocation struct {
								EventAllocationResources struct {
									TotalNodeCount int
									Nodes          []struct {
										Resource struct {
											Name         string
											ResourceType struct {
												Name string
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}
