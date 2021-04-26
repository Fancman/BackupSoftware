package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	//"strconv"

	"github.com/Fancman/BackupSoftware/database"
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	db                = &database.SQLite{}
	source            string
	dest_drive_ksuid  string
	dest_drive_letter string
	//dest_path         string
	//source_path       string
	source_paths []string
	backup_paths []string
	backup_path  string
	archive_name string

	listBackupsCmd = &cobra.Command{
		Use:   "list-backups",
		Short: "List stored backup records",
		RunE: func(cmd *cobra.Command, args []string) error {
			ListBackups()
			/*backups := list_backups(conn)

			if len(backups) > 0 {
				for _, b := range backups {
					fmt.Printf("[id]: %b, [source]: %s, [destinations]: %v", b.id, b.source, b.destinations)
				}
			}*/

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

			ListDrives()

			return nil
		},
	}

	AddDriveCmd = &cobra.Command{
		Use:   "add-drive [drive_letter_identification]",
		Short: "Add drive to db and create .drive",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return errors.New("You have to enter exactly one drive indetification.")
			}

			if (args[0] < "a" || args[0] > "z") && (args[0] < "A" || args[0] > "Z") {
				return errors.New("Typed argument is not an alphabetic letter.")
			}

			if AddDrive(string(args[0])) != "" {
				fmt.Println("Drive was succesfully added.")
			} else {
				fmt.Println("Drive couldnt be added.")
			}

			return nil
		},
	}

	startBackupCmd = &cobra.Command{
		Use:   "start-backup [source id]",
		Short: "Start backup from record in db by its id",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return errors.New("You have to enter exactly one drive indetification.")
			}

			i, err := strconv.ParseInt(args[0], 10, 64)

			if err != nil {
				return errors.New("Entered backup id have to be an integer type")
			}

			BackupFileDir(i)

			//backup := find_backup(db, i)
			//start_backup(backup.id, backup.source, backup.destinations)

			return nil
		},
	}

	startRestoreCmd = &cobra.Command{
		Use:   "start-restore [source id]",
		Short: "Start restore from record in db by its id",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return errors.New("You have to enter exactly one drive indetification.")
			}

			i, err := strconv.ParseInt(args[0], 10, 64)

			if err != nil {
				return errors.New("Entered backup id have to be an integer type")
			}

			RestoreFileDir(i)

			//backup := find_backup(db, i)
			//start_restore(backup.id, backup.source, backup.destinations)

			return nil
		},
	}

	/*createBackupCmd = &cobra.Command{
		Use:   "create-backup -s [source] -d [destination drive letter] -p [path] | optional: -ksuid [drive ksuid]",
		Short: "Create backup record by enetring drive letter or ksuid -s [source] -d [destination drive letter] -p [path] | optional: -ksuid [drive ksuid]",
		RunE: func(cmd *cobra.Command, args []string) error {
			//fmt.Println(source, dest_drive_ksuid, dest_path)
			//insert_backups_db(source, dest_drive_ksuid, dest_drive_letter, dest_path)
			return nil
		},
	}*/

	createBackupCmd = &cobra.Command{
		Use:   "create-backup -s [source path] -d [destination path] -a [archive name]",
		Short: "create-backup -s [source path] -d [destination path] -a [archive name]",
		RunE: func(cmd *cobra.Command, args []string) error {
			//fmt.Println(source_paths)
			CreateSourceBackup(source_paths, backup_paths, archive_name)
			//CreateSourceBackup("C:/Users/tomas/Pictures/Backgrounds", "E:/backup", "test.7z")
			//fmt.Println(source, dest_drive_ksuid, dest_path)
			//insert_backups_db(source, dest_drive_ksuid, dest_drive_letter, dest_path)
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
	BackupDatabase()
	err := db.Fixtures()

	if err != nil {
		fmt.Println("Database tables couldnt be created.")
	}

	rootCmd.AddCommand(listDrivesCmd)
	rootCmd.AddCommand(listBackupsCmd)
	rootCmd.AddCommand(AddDriveCmd)
	rootCmd.AddCommand(createBackupCmd)
	//rootCmd.AddCommand(createBackupCmdTest)

	//createBackupCmd.Flags().StringSlice("source", source_paths, "sources paths")
	createBackupCmd.Flags().StringArrayVarP(&source_paths, "source", "s", make([]string, 0), "sources paths")
	createBackupCmd.Flags().StringArrayVarP(&backup_paths, "destination", "d", make([]string, 0), "destination path")
	//createBackupCmd.Flags().StringVarP(&source_path, "source", "s", "", "source path")
	//createBackupCmd.Flags().StringVarP(&backup_path, "backup", "d", "", "destination path")
	createBackupCmd.Flags().StringVarP(&archive_name, "archive", "a", "", "archive name")

	//BackupFileDir(10)

	//RestoreFileDir(10)

	/*createBackupCmd.Flags().StringVarP(&source, "source", "s", "", "source path")
	createBackupCmd.Flags().StringVarP(&dest_drive_ksuid, "drive_ksuid", "k", "", "destination drive ksuid")
	createBackupCmd.Flags().StringVarP(&dest_drive_letter, "destination drive letter", "d", "", "destination drive letter")
	createBackupCmd.Flags().StringVarP(&dest_path, "additional path", "p", "", "additional path")*/

	rootCmd.AddCommand(startBackupCmd)
	rootCmd.AddCommand(startRestoreCmd)

	//singleCmd.Flags().StringVarP(&options.File, "file", "f", "", "file containing urls. use - for stdin")
	//rootCmd.AddCommand(rootCmd)
}
