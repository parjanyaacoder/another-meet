package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

var (
	successIcon = color.GreenString("✓")
	errorIcon   = color.RedString("✗")
	infoIcon    = color.BlueString("ℹ")
	warnIcon    = color.YellowString("⚠")
)

// Success prints a success message to stderr
func Success(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s %s\n", successIcon, fmt.Sprintf(msg, args...))
}

// Error prints an error message to stderr
func Error(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s %s\n", errorIcon, fmt.Sprintf(msg, args...))
}

// Info prints an info message to stderr
func Info(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s %s\n", infoIcon, fmt.Sprintf(msg, args...))
}

// Warn prints a warning message to stderr
func Warn(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s %s\n", warnIcon, fmt.Sprintf(msg, args...))
}

// Detail prints an indented detail line to stderr
func Detail(key, value string) {
	fmt.Fprintf(os.Stderr, "  %s  %s\n", color.New(color.Faint).Sprint(key+":"), value)
}

// PrintJSON outputs structured data as JSON to stdout
func PrintJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// IsJSON returns true if --json flag is set
func IsJSON() bool {
	return viper.GetBool("json")
}

// Table prints a simple table to stdout
func Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		Info("No results found.")
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Build format string
	headerColor := color.New(color.Bold, color.FgHiWhite)
	borderColor := color.New(color.Faint)

	// Top border
	borderParts := make([]string, len(widths))
	for i, w := range widths {
		borderParts[i] = strings.Repeat("─", w+2)
	}
	borderColor.Fprintf(os.Stdout, "┌%s┐\n", strings.Join(borderParts, "┬"))

	// Headers
	fmt.Fprint(os.Stdout, borderColor.Sprint("│"))
	for i, h := range headers {
		fmt.Fprintf(os.Stdout, " %s%s ", headerColor.Sprint(h), strings.Repeat(" ", widths[i]-len(h)))
		fmt.Fprint(os.Stdout, borderColor.Sprint("│"))
	}
	fmt.Fprintln(os.Stdout)

	// Header separator
	borderColor.Fprintf(os.Stdout, "├%s┤\n", strings.Join(borderParts, "┼"))

	// Rows
	for _, row := range rows {
		fmt.Fprint(os.Stdout, borderColor.Sprint("│"))
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(os.Stdout, " %s%s ", cell, strings.Repeat(" ", widths[i]-len(cell)))
				fmt.Fprint(os.Stdout, borderColor.Sprint("│"))
			}
		}
		fmt.Fprintln(os.Stdout)
	}

	// Bottom border
	borderColor.Fprintf(os.Stdout, "└%s┘\n", strings.Join(borderParts, "┴"))
}

// MeetingCard displays a meeting creation result
func MeetingCard(title, meetLink, startTime, endTime string, attendees []string) {
	titleColor := color.New(color.Bold, color.FgHiCyan)
	linkColor := color.New(color.Bold, color.FgHiGreen)
	dimColor := color.New(color.Faint)

	fmt.Fprintln(os.Stderr)
	Success("Meeting created!")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "  %s  %s\n", dimColor.Sprint("Title:"), titleColor.Sprint(title))
	fmt.Fprintf(os.Stderr, "  %s   %s\n", dimColor.Sprint("Meet:"), linkColor.Sprint(meetLink))
	fmt.Fprintf(os.Stderr, "  %s   %s — %s\n", dimColor.Sprint("Time:"), startTime, endTime)
	if len(attendees) > 0 {
		fmt.Fprintf(os.Stderr, "  %s  %s\n", dimColor.Sprint("Invited:"), strings.Join(attendees, ", "))
	}
	fmt.Fprintln(os.Stderr)
}
