package cmd

import (
	"context"
	"time"

	cal "github.com/parjanyaacoder/another-meet/internal/calendar"
	"github.com/parjanyaacoder/another-meet/internal/config"
	"github.com/parjanyaacoder/another-meet/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List upcoming meetings",
	Long:  `List upcoming calendar events. By default shows today's meetings.`,
	Example: `  # List today's meetings
  another-meet list

  # List meetings for a specific date range
  another-meet list --from "2026-06-10" --to "2026-06-12"

  # List only meetings with Google Meet links
  another-meet list --has-meet

  # Show more results
  another-meet list --max 50

  # Output as JSON
  another-meet list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")
		maxResults, _ := cmd.Flags().GetInt64("max")
		hasMeet, _ := cmd.Flags().GetBool("has-meet")
		calendarID, _ := cmd.Flags().GetString("calendar")

		loc := config.Timezone()
		now := time.Now().In(loc)

		// Parse "from" date
		var timeMin time.Time
		if fromStr != "" {
			t, err := parseDate(fromStr, loc)
			if err != nil {
				return err
			}
			timeMin = t
		} else {
			// Default: start of today
			timeMin = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		}

		// Parse "to" date
		var timeMax time.Time
		if toStr != "" {
			t, err := parseDate(toStr, loc)
			if err != nil {
				return err
			}
			// End of that day
			timeMax = t.Add(24*time.Hour - time.Second)
		} else {
			// Default: end of today
			timeMax = timeMin.Add(24*time.Hour - time.Second)
		}

		events, err := cal.ListEvents(ctx, timeMin, timeMax, maxResults, calendarID, hasMeet)
		if err != nil {
			return err
		}

		// JSON output
		if ui.IsJSON() {
			return ui.PrintJSON(events)
		}

		if len(events) == 0 {
			ui.Info("No meetings found for the selected period.")
			return nil
		}

		// Build table
		headers := []string{"TITLE", "TIME", "MEET LINK"}
		var rows [][]string
		for _, e := range events {
			startDisplay := formatEventTime(e.Start, loc)
			endDisplay := formatEventTime(e.End, loc)
			timeRange := startDisplay + " — " + endDisplay

			meetLink := e.MeetLink
			if meetLink == "" {
				meetLink = "—"
			} else if len(meetLink) > 35 {
				meetLink = meetLink[len(meetLink)-30:]
			}

			title := e.Title
			if len(title) > 30 {
				title = title[:27] + "..."
			}

			rows = append(rows, []string{title, timeRange, meetLink})
		}

		ui.Table(headers, rows)
		return nil
	},
}

func parseDate(s string, loc *time.Location) (time.Time, error) {
	// Try full datetime first
	for _, layout := range []string{
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
	} {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t, nil
		}
	}

	// Try relative dates
	now := time.Now().In(loc)
	switch s {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc), nil
	case "tomorrow":
		tomorrow := now.Add(24 * time.Hour)
		return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, loc), nil
	}

	return time.Time{}, nil
}

func formatEventTime(dt string, loc *time.Location) string {
	// Try RFC3339
	if t, err := time.Parse(time.RFC3339, dt); err == nil {
		return t.In(loc).Format("3:04 PM")
	}
	// Try date-only
	if t, err := time.Parse("2006-01-02", dt); err == nil {
		return t.Format("Jan 02")
	}
	return dt
}

func init() {
	listCmd.Flags().String("from", "", "start date (YYYY-MM-DD or 'today'/'tomorrow')")
	listCmd.Flags().String("to", "", "end date (YYYY-MM-DD)")
	listCmd.Flags().Int64("max", 20, "maximum number of results")
	listCmd.Flags().Bool("has-meet", false, "only show events with Google Meet links")
	listCmd.Flags().String("calendar", "", "calendar ID (default: primary)")

	rootCmd.AddCommand(listCmd)
}
