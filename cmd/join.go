package cmd

import (
	"context"
	"fmt"

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

  # Join a specific meeting by event ID
  another-meet join --id <event-id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		eventID, _ := cmd.Flags().GetString("id")
		calendarID, _ := cmd.Flags().GetString("calendar")

		var event *cal.Event
		var err error

		if eventID != "" {
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
