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
	db = open_conn()

	listBackupsCmd = &cobra.Command{
		Use:   "list-backups",
		Short: "List stored backup records",
		RunE: func(cmd *cobra.Command, args []string) error {
			backups := list_backups(db)

			if len(backups) > 0 {
				for _, b := range backups {
					fmt.Printf("[id]: %b, [source]: %s, [destinations]: %v", b.id, b.source, b.destinations)
				}
			}

			return nil
		},
	}

	listDrivesCmd = &cobra.Command{
		Use:   "list-drives",
		Short: "List available drives",
		RunE: func(cmd *cobra.Command, args []string) error {
			/*if len(args) == 0 {
				return errors.New("You have to enter target of action.")
			}*/

			list_drives()

			return nil
		},
	}

	rootCmd = &cobra.Command{
		Use:   "Backupsoft",
		Short: "Simple 7z backup utility",
		Long:  `Backupsoft is a distributed backup utility using 7z as its tool for archiving.`,
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
	rootCmd.AddCommand(listDrivesCmd)
	rootCmd.AddCommand(listBackupsCmd)
	rootCmd.AddCommand(singleCmd)
	//singleCmd.Flags().StringVarP(&options.File, "file", "f", "", "file containing urls. use - for stdin")
	//rootCmd.AddCommand(rootCmd)
}
