package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	cal "github.com/parjanyaacoder/another-meet/internal/calendar"
	"github.com/parjanyaacoder/another-meet/internal/config"
	"github.com/parjanyaacoder/another-meet/internal/ui"
	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Schedule a meeting for a future time",
	Long:  `Schedule a Google Calendar event with a Google Meet link at a specific future time.`,
	Example: `  # Schedule a meeting at a specific time
  another-meet schedule --title "Design Review" --at "2026-06-10 14:00" --duration 45m

  # Schedule with attendees
  another-meet schedule -t "Weekly Sync" --at "tomorrow 10:00" -d 30m -a "team@company.com"

  # Schedule for tomorrow
  another-meet schedule -t "Standup" --at "tomorrow 09:30"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		title, _ := cmd.Flags().GetString("title")
		atStr, _ := cmd.Flags().GetString("at")
		duration, _ := cmd.Flags().GetString("duration")
		attendeesStr, _ := cmd.Flags().GetString("attendees")
		description, _ := cmd.Flags().GetString("description")
		openBrowser, _ := cmd.Flags().GetBool("open")
		calendarID, _ := cmd.Flags().GetString("calendar")

		if atStr == "" {
			return fmt.Errorf("--at is required. Specify when the meeting should start (e.g., '2026-06-10 14:00' or 'tomorrow 10:00')")
		}

		// Parse duration
		dur, err := time.ParseDuration(duration)
		if err != nil {
			return fmt.Errorf("invalid duration %q: %w", duration, err)
		}

		// Parse the scheduled time
		loc := config.Timezone()
		start, err := parseScheduleTime(atStr, loc)
		if err != nil {
			return fmt.Errorf("invalid time %q: %w", atStr, err)
		}

		end := start.Add(dur)

		// Parse attendees
		var attendees []string
		if attendeesStr != "" {
			for _, email := range strings.Split(attendeesStr, ",") {
				email = strings.TrimSpace(email)
				if email != "" {
					attendees = append(attendees, email)
				}
			}
		}

		params := cal.CreateEventParams{
			Title:       title,
			Description: description,
			Start:       start,
			End:         end,
			Attendees:   attendees,
			WithMeet:    true,
			CalendarID:  calendarID,
		}

		event, err := cal.CreateEvent(ctx, params)
		if err != nil {
			return err
		}

		// JSON output
		if ui.IsJSON() {
			return ui.PrintJSON(event)
		}

		startTime := start.In(loc).Format("Mon, Jan 2 at 3:04 PM")
		endTime := end.In(loc).Format("3:04 PM MST")

		ui.MeetingCard(event.Title, event.MeetLink, startTime, endTime, attendees)

		if openBrowser && event.MeetLink != "" {
			openURL(event.MeetLink)
		}

		return nil
	},
}

func parseScheduleTime(s string, loc *time.Location) (time.Time, error) {
	now := time.Now().In(loc)

	// Handle "tomorrow HH:MM"
	if strings.HasPrefix(s, "tomorrow") {
		timeStr := strings.TrimSpace(strings.TrimPrefix(s, "tomorrow"))
		tomorrow := now.Add(24 * time.Hour)

		if timeStr == "" {
			return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 0, 0, 0, loc), nil
		}

		t, err := time.Parse("15:04", timeStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid time format, use HH:MM (e.g., '14:30')")
		}
		return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), t.Hour(), t.Minute(), 0, 0, loc), nil
	}

	// Handle "today HH:MM"
	if strings.HasPrefix(s, "today") {
		timeStr := strings.TrimSpace(strings.TrimPrefix(s, "today"))
		if timeStr == "" {
			return now, nil
		}

		t, err := time.Parse("15:04", timeStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid time format, use HH:MM (e.g., '14:30')")
		}
		return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc), nil
	}

	// Try full datetime formats
	for _, layout := range []string{
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"Jan 2 15:04",
		"January 2 15:04",
	} {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			// If year is zero, use current year
			if t.Year() == 0 {
				t = time.Date(now.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time. Use formats like '2026-06-10 14:00', 'tomorrow 10:00', or 'today 15:30'")
}

func init() {
	scheduleCmd.Flags().StringP("title", "t", "Scheduled Meeting", "meeting title")
	scheduleCmd.Flags().String("at", "", "when to schedule (e.g., '2026-06-10 14:00', 'tomorrow 10:00') [required]")
	scheduleCmd.Flags().StringP("duration", "d", "30m", "meeting duration")
	scheduleCmd.Flags().StringP("attendees", "a", "", "comma-separated email addresses")
	scheduleCmd.Flags().String("description", "", "meeting description")
	scheduleCmd.Flags().BoolP("open", "o", false, "open Meet link in browser")
	scheduleCmd.Flags().String("calendar", "", "calendar ID (default: primary)")

	rootCmd.AddCommand(scheduleCmd)
}
