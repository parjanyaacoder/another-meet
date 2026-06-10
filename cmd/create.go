package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	cal "github.com/parjanyaacoder/another-meet/internal/calendar"
	"github.com/parjanyaacoder/another-meet/internal/config"
	"github.com/parjanyaacoder/another-meet/internal/ui"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new meeting with Google Meet link",
	Long: `Create a new Google Calendar event with a Google Meet link.
If no options are provided, creates an instant 30-minute meeting.`,
	Example: `  # Create an instant meeting
  another-meet create

  # Create with a custom title and duration
  another-meet create --title "Sprint Planning" --duration 1h

  # Create and invite attendees
  another-meet create -t "Design Review" -d 45m -a "alice@company.com,bob@company.com"

  # Create and immediately open in browser
  another-meet create --open`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		title, _ := cmd.Flags().GetString("title")
		duration, _ := cmd.Flags().GetString("duration")
		attendeesStr, _ := cmd.Flags().GetString("attendees")
		description, _ := cmd.Flags().GetString("description")
		openBrowser, _ := cmd.Flags().GetBool("open")
		calendarID, _ := cmd.Flags().GetString("calendar")
		noMeet, _ := cmd.Flags().GetBool("no-meet")

		// Parse duration
		dur, err := time.ParseDuration(duration)
		if err != nil {
			return fmt.Errorf("invalid duration %q: %w", duration, err)
		}

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

		start := time.Now()
		end := start.Add(dur)

		params := cal.CreateEventParams{
			Title:       title,
			Description: description,
			Start:       start,
			End:         end,
			Attendees:   attendees,
			WithMeet:    !noMeet,
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

		// Format times for display
		loc := config.Timezone()
		startTime := start.In(loc).Format("3:04 PM")
		endTime := end.In(loc).Format("3:04 PM MST")

		ui.MeetingCard(event.Title, event.MeetLink, startTime, endTime, attendees)

		// Open in browser if requested
		if openBrowser && event.MeetLink != "" {
			openURL(event.MeetLink)
		}

		return nil
	},
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}

func init() {
	createCmd.Flags().StringP("title", "t", "Quick Meeting", "meeting title")
	createCmd.Flags().StringP("duration", "d", "30m", "meeting duration (e.g., 30m, 1h, 1h30m)")
	createCmd.Flags().StringP("attendees", "a", "", "comma-separated email addresses")
	createCmd.Flags().String("description", "", "meeting description")
	createCmd.Flags().BoolP("open", "o", false, "open Meet link in browser after creation")
	createCmd.Flags().String("calendar", "", "calendar ID (default: primary)")
	createCmd.Flags().Bool("no-meet", false, "create event without Google Meet link")

	rootCmd.AddCommand(createCmd)
}
