package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/parjanyaacoder/another-meet/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func printWelcomeScreen(cmd *cobra.Command) {
	logo := []string{
		`  ███████████████    `,
		`  ██           ██ ██ `,
		`  ██  M E E T  █████ `,
		`  ██           ██ ██ `,
		`  ███████████████    `,
	}

	infoText := []string{
		color.New(color.FgCyan, color.Bold).Sprint("another-meet") + " - Manage Google Meet from your terminal",
		"--------------------------------------------------",
		color.New(color.FgYellow, color.Bold).Sprint("Get started:"),
		"  " + color.New(color.FgGreen).Sprint("another-meet auth login") + "     # Authenticate with Google",
		"  " + color.New(color.FgGreen).Sprint("another-meet create") + "         # Create an instant meeting",
		"  " + color.New(color.FgGreen).Sprint("another-meet list") + "           # List upcoming meetings",
		"  " + color.New(color.FgGreen).Sprint("another-meet join") + "           # Join the next meeting",
		"",
		"Run " + color.New(color.FgHiBlue).Sprint("another-meet --help") + " for a full list of commands.",
	}

	logoColors := []*color.Color{
		color.New(color.FgHiBlue),
		color.New(color.FgHiBlue),
		color.New(color.FgBlue),
		color.New(color.FgBlue),
		color.New(color.FgCyan),
	}

	maxLines := len(logo)
	if len(infoText) > maxLines {
		maxLines = len(infoText)
	}

	fmt.Println()
	for i := 0; i < maxLines; i++ {
		logoLine := "                     " // 21 spaces padding
		if i < len(logo) {
			logoLine = logoColors[i].Sprint(logo[i])
		}

		infoLine := ""
		if i < len(infoText) {
			infoLine = infoText[i]
		}

		fmt.Printf("  %s  %s\n", logoLine, infoLine)
	}
	fmt.Println()
}

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "another-meet",
	Short: "Manage meetings from your terminal",
	Long: `another-meet is a CLI tool for creating, joining, scheduling,
and managing Google Meet meetings directly from your terminal.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, args []string) {
		printWelcomeScreen(cmd)
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.another-meet/config.yaml)")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")
	rootCmd.PersistentFlags().Bool("json", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	_ = viper.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no-color"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		configDir := config.Dir()
		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("ANOTHER_MEET")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Warning: error reading config: %v\n", err)
		}
	}
}
