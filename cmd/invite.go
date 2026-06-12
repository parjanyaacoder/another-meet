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

var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Add attendees to an existing meeting",
	Long:  `Invite people to an existing Google Calendar meeting by adding their email addresses.`,
	Example: `  # Invite to a specific meeting by Calendar Event ID
  another-meet invite --id <event-id> --attendees "charlie@company.com"

  # Invite directly using a Meet code or URL
  another-meet invite --id abc-defg-hij -a "dave@company.com"

  # Invite to the next upcoming meeting
  another-meet invite --next --attendees "charlie@company.com,dave@company.com"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		eventID, _ := cmd.Flags().GetString("id")
		attendeesStr, _ := cmd.Flags().GetString("attendees")
		next, _ := cmd.Flags().GetBool("next")
		calendarID, _ := cmd.Flags().GetString("calendar")

		if attendeesStr == "" {
			return fmt.Errorf("--attendees is required. Provide comma-separated email addresses")
		}

		// Parse attendees
		var emails []string
		for _, email := range strings.Split(attendeesStr, ",") {
			email = strings.TrimSpace(email)
			if email != "" {
				emails = append(emails, email)
			}
		}

		if len(emails) == 0 {
			return fmt.Errorf("no valid email addresses provided")
		}

		// Get the event ID
		if eventID == "" && next {
			event, err := cal.GetNextMeeting(ctx, calendarID)
			if err != nil {
				return err
			}
			eventID = event.ID
		}

		if eventID == "" {
			return fmt.Errorf("specify --id <event-id> or --next to invite to the next meeting")
		}

		// Check if the user passed a raw Meet code or URL
		meetCodeRegex := regexp.MustCompile(`(?i)([a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{3})`)
		if strings.Contains(eventID, "meet.google.com/") || meetCodeRegex.MatchString(eventID) {
			match := meetCodeRegex.FindString(eventID)
			if match == "" {
				return fmt.Errorf("invalid Google Meet format. Expected format: abc-defg-hij")
			}
			
			ui.Info("Searching your calendar for Meet code: %s...", match)
			ev, err := cal.FindEventByMeetCode(ctx, match, calendarID)
			if err != nil {
				return err
			}
			eventID = ev.ID
		}

		event, err := cal.AddAttendees(ctx, eventID, emails, calendarID)
		if err != nil {
			return err
		}

		// JSON output
		if ui.IsJSON() {
			return ui.PrintJSON(event)
		}

		ui.Success("Invited %d attendee(s) to %q", len(emails), event.Title)
		for _, email := range emails {
			ui.Detail("  →", email)
		}

		return nil
	},
}

func init() {
	inviteCmd.Flags().String("id", "", "event ID to invite attendees to")
	inviteCmd.Flags().StringP("attendees", "a", "", "comma-separated email addresses [required]")
	inviteCmd.Flags().Bool("next", false, "invite to the next upcoming meeting")
	inviteCmd.Flags().String("calendar", "", "calendar ID (default: primary)")

	rootCmd.AddCommand(inviteCmd)
}
