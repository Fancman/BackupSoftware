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
	archive_id   int64
	drive_ksuid  string
	source_id    int64

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

	addDriveCmd = &cobra.Command{
		Use:   "add-drive [drive_letter_identification] -n [Custom drive name]",
		Short: "Add drive to db with optional custom name and create .drive file",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return errors.New("You have to enter exactly one drive identification.")
			}

			if (args[0] < "a" || args[0] > "z") && (args[0] < "A" || args[0] > "Z") {
				return errors.New("Typed argument is not an alphabetic letter.")
			}

			drive_name, err := cmd.Flags().GetString("drive-name")

			if err != nil {
				return err
			}

			if AddDrive(string(args[0]), drive_name) != "" {
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
		Use:   "start-restore [source id] -b [backup paths]",
		Short: "Start restore from record in db by its id",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return errors.New("You have to enter exactly one drive indetification.")
			}

			i, err := strconv.ParseInt(args[0], 10, 64)

			if err != nil {
				return errors.New("Entered backup id have to be an integer type")
			}

			RestoreFileDir(i, backup_paths)

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

	removeSourceCmd = &cobra.Command{
		Use:   "remove-source -s [source id]",
		Short: "remove-source -s [source id]",
		RunE: func(cmd *cobra.Command, args []string) error {

			source_id, err := cmd.Flags().GetInt64("source-id")

			if err != nil {
				return err
			}

			RemoveSource(source_id)

			return nil
		},
	}

	removeDestinationCmd = &cobra.Command{
		Use:   "remove-destination -a [archive id] -d [drive ksuid]",
		Short: "remove-destination -a [archive id] -d [drive ksuid]",
		RunE: func(cmd *cobra.Command, args []string) error {

			archive_id, err_1 := cmd.Flags().GetInt64("archive-id")
			drive_ksuid, err_2 := cmd.Flags().GetString("drive-ksuid")

			if err_1 != nil {
				return err_1
			}

			if err_2 != nil {
				return err_2
			}

			if archive_id != 0 && drive_ksuid != "" {
				RemoveDestination(archive_id, drive_ksuid)
			}

			destination_path, err := cmd.Flags().GetString("dest-path")

			if err != nil {
				return err
			}

			if destination_path != "" {
				RemoveDestinationByPath(destination_path)
			}

			return nil
		},
	}

	//insert_backups_db(source string, dest_drive_ksuid string, path string)

	rootCmd = &cobra.Command{
		Use:   "Backupsoft",
		Short: "Simple 7z backup utility",
		Long:  `Backupsoft is a distributed backup utility using 7z as its tool for archiving.`,
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
	rootCmd.AddCommand(addDriveCmd)
	rootCmd.AddCommand(createBackupCmd)
	rootCmd.AddCommand(removeSourceCmd)
	rootCmd.AddCommand(removeDestinationCmd)
	rootCmd.AddCommand(startBackupCmd)
	rootCmd.AddCommand(startRestoreCmd)

	addDriveCmd.Flags().StringP("drive-name", "n", "", "Drive name")

	removeSourceCmd.Flags().Int64P("source-id", "s", 0, "Source ID")

	removeDestinationCmd.Flags().Int64P("archive-id", "a", 0, "Archive ID")
	removeDestinationCmd.Flags().StringP("drive-ksuid", "d", "", "Drive Ksuid")
	removeDestinationCmd.Flags().StringP("dest-path", "p", "", "Destination path")

	//rootCmd.AddCommand(createBackupCmdTest)

	//createBackupCmd.Flags().StringSlice("source", source_paths, "sources paths")
	createBackupCmd.Flags().StringArrayVarP(&source_paths, "source", "s", make([]string, 0), "sources paths")
	createBackupCmd.Flags().StringArrayVarP(&backup_paths, "destination", "d", make([]string, 0), "destination path")
	createBackupCmd.Flags().StringVarP(&archive_name, "archive", "a", "", "archive name")

	startRestoreCmd.Flags().StringArrayVarP(&backup_paths, "backup", "b", make([]string, 0), "backup paths")

	//BackupFileDir(10)

	//RestoreFileDir(10)

	/*createBackupCmd.Flags().StringVarP(&source, "source", "s", "", "source path")
	createBackupCmd.Flags().StringVarP(&dest_drive_ksuid, "drive_ksuid", "k", "", "destination drive ksuid")
	createBackupCmd.Flags().StringVarP(&dest_drive_letter, "destination drive letter", "d", "", "destination drive letter")
	createBackupCmd.Flags().StringVarP(&dest_path, "additional path", "p", "", "additional path")*/

	//singleCmd.Flags().StringVarP(&options.File, "file", "f", "", "file containing urls. use - for stdin")
	//rootCmd.AddCommand(rootCmd)
}
