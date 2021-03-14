package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	db               = open_conn()
	source           string
	dest_drive_ksuid string
	dest_path        string

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

	addDriveCmd = &cobra.Command{
		Use:   "add-drive [drive_letter_identification]",
		Short: "Add drive to db and create .drive",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return errors.New("You have to enter exactly one drive indetification.")
			}

			add_drive((args[0]))

			return nil
		},
	}

	startBackupCmd = &cobra.Command{
		Use:   "start-backup [backup id]",
		Short: "Start backup from record in db by its id",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return errors.New("You have to enter exactly one drive indetification.")
			}

			i, err := strconv.Atoi(args[0])

			if err != nil {
				return errors.New("Entered backup id have to be an integer type")
			}

			backup := find_backup(db, i)
			start_backup(backup.id, backup.source, backup.destinations)

			return nil
		},
	}

	startRestoreCmd = &cobra.Command{
		Use:   "start-restore [backup id]",
		Short: "Start restore from record in db by its id",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return errors.New("You have to enter exactly one drive indetification.")
			}

			i, err := strconv.Atoi(args[0])

			if err != nil {
				return errors.New("Entered backup id have to be an integer type")
			}

			backup := find_backup(db, i)
			start_restore(backup.id, backup.source, backup.destinations)

			return nil
		},
	}

	createBackupCmd = &cobra.Command{
		Use:   "create-backup -s [source] -d [destination drive] -p [path]",
		Short: "Create backup record -s [source] -d [destination drive] -p [path]",
		RunE: func(cmd *cobra.Command, args []string) error {
			//fmt.Println(source, dest_drive_ksuid, dest_path)
			insert_backups_db(source, dest_drive_ksuid, dest_path)
			return nil
		},
	}

	//insert_backups_db(source string, dest_drive_ksuid string, path string)

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
	rootCmd.AddCommand(addDriveCmd)
	rootCmd.AddCommand(createBackupCmd)

	createBackupCmd.Flags().StringVarP(&source, "source", "s", "", "source path")
	createBackupCmd.Flags().StringVarP(&dest_drive_ksuid, "drive_ksuid", "d", "", "destination drive ksuid")
	createBackupCmd.Flags().StringVarP(&dest_path, "additional path", "p", "", "additional path")

	rootCmd.AddCommand(startBackupCmd)
	rootCmd.AddCommand(startRestoreCmd)

	//singleCmd.Flags().StringVarP(&options.File, "file", "f", "", "file containing urls. use - for stdin")
	//rootCmd.AddCommand(rootCmd)
}
