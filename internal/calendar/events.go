package cal

import (
	"context"
	"fmt"
	"time"

	"github.com/parjanyaacoder/another-meet/internal/config"
	"google.golang.org/api/calendar/v3"
)

// Event represents a simplified calendar event for display
type Event struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Start       string   `json:"start"`
	End         string   `json:"end"`
	MeetLink    string   `json:"meet_link,omitempty"`
	HtmlLink    string   `json:"html_link"`
	Attendees   []string `json:"attendees,omitempty"`
	Description string   `json:"description,omitempty"`
}

// CreateEventParams holds parameters for creating a new event
type CreateEventParams struct {
	Title       string
	Description string
	Start       time.Time
	End         time.Time
	Attendees   []string
	WithMeet    bool
	CalendarID  string
}

// CreateEvent creates a new calendar event, optionally with a Google Meet link.
func CreateEvent(ctx context.Context, params CreateEventParams) (*Event, error) {
	srv, err := NewService(ctx)
	if err != nil {
		return nil, err
	}

	if params.CalendarID == "" {
		params.CalendarID = config.DefaultCalendar()
	}

	event := &calendar.Event{
		Summary:     params.Title,
		Description: params.Description,
		Start: &calendar.EventDateTime{
			DateTime: params.Start.Format(time.RFC3339),
			TimeZone: config.TimezoneName(),
		},
		End: &calendar.EventDateTime{
			DateTime: params.End.Format(time.RFC3339),
			TimeZone: config.TimezoneName(),
		},
	}

	// Add attendees
	if len(params.Attendees) > 0 {
		attendees := make([]*calendar.EventAttendee, len(params.Attendees))
		for i, email := range params.Attendees {
			attendees[i] = &calendar.EventAttendee{Email: email}
		}
		event.Attendees = attendees
	}

	// Add Google Meet conference data
	if params.WithMeet {
		event.ConferenceData = NewMeetConference()
	}

	// Insert the event
	call := srv.Events.Insert(params.CalendarID, event)
	if params.WithMeet {
		call = call.ConferenceDataVersion(1)
	}
	call = call.SendUpdates("all") // Notify attendees

	created, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return eventToResult(created), nil
}

// ListEvents retrieves upcoming events from the calendar.
func ListEvents(ctx context.Context, timeMin, timeMax time.Time, maxResults int64, calendarID string, meetOnly bool) ([]*Event, error) {
	srv, err := NewService(ctx)
	if err != nil {
		return nil, err
	}

	if calendarID == "" {
		calendarID = config.DefaultCalendar()
	}
	if maxResults == 0 {
		maxResults = 20
	}

	call := srv.Events.List(calendarID).
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		MaxResults(maxResults).
		SingleEvents(true).
		OrderBy("startTime")

	events, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	var results []*Event
	for _, item := range events.Items {
		e := eventToResult(item)
		if meetOnly && e.MeetLink == "" {
			continue
		}
		results = append(results, e)
	}

	return results, nil
}

// GetEvent retrieves a single event by ID.
func GetEvent(ctx context.Context, eventID, calendarID string) (*Event, error) {
	srv, err := NewService(ctx)
	if err != nil {
		return nil, err
	}

	if calendarID == "" {
		calendarID = config.DefaultCalendar()
	}

	event, err := srv.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return eventToResult(event), nil
}

// AddAttendees adds attendees to an existing event.
func AddAttendees(ctx context.Context, eventID string, emails []string, calendarID string) (*Event, error) {
	srv, err := NewService(ctx)
	if err != nil {
		return nil, err
	}

	if calendarID == "" {
		calendarID = config.DefaultCalendar()
	}

	// Get the existing event
	event, err := srv.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Add new attendees
	for _, email := range emails {
		event.Attendees = append(event.Attendees, &calendar.EventAttendee{Email: email})
	}

	// Update the event
	updated, err := srv.Events.Update(calendarID, eventID, event).
		SendUpdates("all").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return eventToResult(updated), nil
}

// GetNextMeeting returns the next upcoming event with a Google Meet link.
func GetNextMeeting(ctx context.Context, calendarID string) (*Event, error) {
	now := time.Now()
	end := now.Add(24 * time.Hour)

	events, err := ListEvents(ctx, now, end, 10, calendarID, true)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no upcoming meetings with Google Meet links found in the next 24 hours")
	}

	return events[0], nil
}

func eventToResult(item *calendar.Event) *Event {
	e := &Event{
		ID:          item.Id,
		Title:       item.Summary,
		HtmlLink:    item.HtmlLink,
		Description: item.Description,
	}

	// Parse start time
	if item.Start != nil {
		if item.Start.DateTime != "" {
			e.Start = item.Start.DateTime
		} else {
			e.Start = item.Start.Date
		}
	}

	// Parse end time
	if item.End != nil {
		if item.End.DateTime != "" {
			e.End = item.End.DateTime
		} else {
			e.End = item.End.Date
		}
	}

	// Extract Meet link
	if item.HangoutLink != "" {
		e.MeetLink = item.HangoutLink
	} else if item.ConferenceData != nil {
		for _, ep := range item.ConferenceData.EntryPoints {
			if ep.EntryPointType == "video" {
				e.MeetLink = ep.Uri
				break
			}
		}
	}

	// Extract attendee emails
	for _, a := range item.Attendees {
		e.Attendees = append(e.Attendees, a.Email)
	}

	return e
}
