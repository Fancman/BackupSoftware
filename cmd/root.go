package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func someFunc() error {
	fmt.Println("balalala")
	return nil
}

var (
	// Used for flags.
	cfgFile     string
	userLicense string

	tryCmd = &cobra.Command{
		Use:   "try",
		Short: "Try and possibly fail at something",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := someFunc(); err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd = &cobra.Command{
		Use:   "backp",
		Short: "Simple 7z backup utility",
		Long: `backp is a distributed backup utility
		using 7z as its tool for archiving.`,
	}
	singleCmd = &cobra.Command{
		Use:   "single [URL]",
		Short: "Take a screenshot of a single URL",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("test")
		},
	}
)

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(tryCmd)
	rootCmd.AddCommand(singleCmd)
	//rootCmd.AddCommand(rootCmd)
}
