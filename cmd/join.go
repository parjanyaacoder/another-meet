package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	cal "github.com/parjanyaacoder/another-meet/internal/calendar"
	"github.com/parjanyaacoder/another-meet/internal/ui"
	"github.com/spf13/cobra"
)

var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Open a meeting in your browser",
	Long:  `Join an upcoming Google Meet meeting by opening its link in your default browser.`,
	Example: `  # Join the next upcoming meeting
  another-meet join

  # Join a specific meeting by Calendar Event ID
  another-meet join --id <event-id>

  # Join directly using a Meet code or URL
  another-meet join --id abc-defg-hij
  another-meet join --id https://meet.google.com/abc-defg-hij`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		eventID, _ := cmd.Flags().GetString("id")
		calendarID, _ := cmd.Flags().GetString("calendar")

		var event *cal.Event
		var err error

		if eventID != "" {
			// Check if the user passed a raw Meet code or URL instead of a Calendar Event ID
			meetCodeRegex := regexp.MustCompile(`(?i)([a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{3})`)
			if strings.Contains(eventID, "meet.google.com/") || meetCodeRegex.MatchString(eventID) {
				match := meetCodeRegex.FindString(eventID)
				if match == "" {
					return fmt.Errorf("invalid Google Meet format. Expected format: abc-defg-hij")
				}
				meetLink := fmt.Sprintf("https://meet.google.com/%s", match)
				
				if ui.IsJSON() {
					return ui.PrintJSON(map[string]string{
						"title":     "Direct Meet Link",
						"meet_link": meetLink,
					})
				}
				ui.Info("Joining direct Meet code: %s", match)
				ui.Success("Opening %s in your browser...", meetLink)
				openURL(meetLink)
				return nil
			}

			// Otherwise, treat it as a Calendar Event ID
			event, err = cal.GetEvent(ctx, eventID, calendarID)
		} else {
			event, err = cal.GetNextMeeting(ctx, calendarID)
		}

		if err != nil {
			return err
		}

		if event.MeetLink == "" {
			return fmt.Errorf("event %q has no Google Meet link", event.Title)
		}

		// JSON output
		if ui.IsJSON() {
			return ui.PrintJSON(map[string]string{
				"title":     event.Title,
				"meet_link": event.MeetLink,
			})
		}

		ui.Info("Joining: %s", event.Title)
		ui.Success("Opening %s in your browser...", event.MeetLink)
		openURL(event.MeetLink)

		return nil
	},
}

func init() {
	joinCmd.Flags().String("id", "", "specific event ID to join")
	joinCmd.Flags().String("calendar", "", "calendar ID (default: primary)")

	rootCmd.AddCommand(joinCmd)
}
