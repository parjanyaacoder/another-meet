package cal

import (
	"context"

	"github.com/parjanyaacoder/another-meet/internal/auth"
	"google.golang.org/api/calendar/v3"
)

// NewService creates an authenticated Google Calendar service.
func NewService(ctx context.Context) (*calendar.Service, error) {
	return auth.NewCalendarService(ctx)
}
