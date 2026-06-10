package cal

import (
	"github.com/google/uuid"
	"google.golang.org/api/calendar/v3"
)

// NewMeetConference creates ConferenceData for a Google Meet link.
// The conference data uses the "hangoutsMeet" solution key, which is the
// correct key for generating Google Meet links via the Calendar API.
func NewMeetConference() *calendar.ConferenceData {
	return &calendar.ConferenceData{
		CreateRequest: &calendar.CreateConferenceRequest{
			RequestId: uuid.New().String(),
			ConferenceSolutionKey: &calendar.ConferenceSolutionKey{
				Type: "hangoutsMeet",
			},
		},
	}
}

// ExtractMeetLink extracts the Google Meet link from an event's conference data.
func ExtractMeetLink(event *calendar.Event) string {
	if event.HangoutLink != "" {
		return event.HangoutLink
	}

	if event.ConferenceData != nil {
		for _, ep := range event.ConferenceData.EntryPoints {
			if ep.EntryPointType == "video" {
				return ep.Uri
			}
		}
	}

	return ""
}
