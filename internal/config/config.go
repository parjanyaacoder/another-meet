package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

const (
	appName = "another-meet"
)

// Dir returns the configuration directory (~/.another-meet/)
func Dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, "."+appName)
}

// EnsureDir creates the config directory if it doesn't exist
func EnsureDir() error {
	return os.MkdirAll(Dir(), 0700)
}

// DefaultCalendar returns the configured default calendar ID
func DefaultCalendar() string {
	cal := viper.GetString("default_calendar")
	if cal == "" {
		return "primary"
	}
	return cal
}

// DefaultDuration returns the configured default meeting duration
func DefaultDuration() time.Duration {
	d := viper.GetString("default_duration")
	if d == "" {
		d = "30m"
	}
	dur, err := time.ParseDuration(d)
	if err != nil {
		return 30 * time.Minute
	}
	return dur
}

// Timezone returns the configured timezone location
func Timezone() *time.Location {
	tz := viper.GetString("timezone")
	if tz == "" {
		return time.Local
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Local
	}
	return loc
}

// TimezoneName returns a valid IANA timezone name for the Google Calendar API.
// time.Local.String() returns "Local" which is invalid for the API,
// so we resolve the actual zone name from the current time.
func TimezoneName() string {
	tz := viper.GetString("timezone")
	if tz != "" {
		return tz
	}
	// Get the IANA name from the system timezone
	name, _ := time.Now().Zone()
	// The zone abbreviation (e.g., "IST") isn't valid for Calendar API either.
	// We need to get the full IANA name from the location.
	loc := time.Local
	if loc == nil {
		return "UTC"
	}
	// time.Local on most systems has the proper IANA name
	locName := loc.String()
	if locName == "Local" || locName == "" {
		// Fallback: use the offset to construct a timezone
		_, offset := time.Now().In(loc).Zone()
		hours := offset / 3600
		mins := (offset % 3600) / 60
		if mins < 0 {
			mins = -mins
		}
		_ = name // suppress unused
		if mins == 0 {
			return fmt.Sprintf("Etc/GMT%+d", -hours)
		}
		// Can't represent fractional offsets with Etc/GMT, use UTC as safe default
		return "UTC"
	}
	return locName
}

// OpenBrowser returns whether to auto-open browser after creating a meeting
func OpenBrowser() bool {
	return viper.GetBool("open_browser")
}

// InitDefaults sets up viper default values
func InitDefaults() {
	viper.SetDefault("default_calendar", "primary")
	viper.SetDefault("default_duration", "30m")
	viper.SetDefault("timezone", "")
	viper.SetDefault("open_browser", false)
}
