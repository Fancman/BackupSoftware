package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	//"strconv"

	"github.com/Fancman/BackupSoftware/database"
	helper "github.com/Fancman/BackupSoftware/helpers"
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	db = &database.SQLite{}

	//source_path       string
	source_paths  []string
	backup_paths  []string
	source_ids    []int64
	archive_names []string

	archive_name   string
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
		Short: "List available drives.",
		RunE: func(cmd *cobra.Command, args []string) error {
			/*if len(args) == 0 {
				return errors.New("You have to enter target of action.")
			}*/

			ListDrives()

			return nil
		},
	}

	ClearAllTablesCmd = &cobra.Command{
		Use:   "clear-tables",
		Short: "Deletes all records from tables.",
		RunE: func(cmd *cobra.Command, args []string) error {

			db.ClearAllTables()

			return nil
		},
	}

	addDriveCmd = &cobra.Command{
		Use:   "add-drive [drive letter] -n [Custom drive name]",
		Short: "Add drive to db with optional custom name and create .drive file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return errors.New("you have to enter exactly one drive identification")
			}

			if (args[0] < "a" || args[0] > "z") && (args[0] < "A" || args[0] > "Z") {
				return errors.New("typed argument is not an alphabetic letter")
			}

			drive_name, err := cmd.Flags().GetString("drive-name")

			if err != nil {
				return err
			}

			AddDrive(string(args[0]), drive_name)

			return nil
		},
	}

	startBackupCmd = &cobra.Command{
		Use:   "start-backup -s [source ids] -a [archive name]",
		Short: "Start backup for record in db by its ids",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(source_ids) == 0 && len(archive_names) == 0 {
				return errors.New("requires atleast one source id or archive name")
			}

			BackupFileDir(source_ids, archive_names)

			//backup := find_backup(db, i)
			//start_backup(backup.id, backup.source, backup.destinations)

			return nil
		},
	}

	startRestoreCmd = &cobra.Command{
		Use:   "start-restore -s [source ids] -b [backup paths]",
		Short: "Start restore from record in db by its id",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(source_ids) == 0 && len(archive_names) == 0 {
				return errors.New("requires atleast one source id or archive name")
			}

			RestoreFileDir(source_ids, archive_names, backup_paths)

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
		Use:   "create-backup -s [source paths] -d [destination paths] -a [archive name]",
		Short: "Create backup record from source and destination paths. Archive name is optional.",
		RunE: func(cmd *cobra.Command, args []string) error {
			//fmt.Println(source_paths)
			if len(source_paths) == 0 {
				return errors.New("requires atleast one source path")
			}

			if len(backup_paths) == 0 {
				return errors.New("requires atleast one destination path")
			}

			var backup_drives []string

			for _, backup_path := range backup_paths {
				backup_letter := strings.ReplaceAll(filepath.VolumeName(backup_path), ":", "")

				if helper.FindElm(backup_drives, backup_letter) {
					return errors.New("destinations have to be on different drives")
				}

				backup_drives = append(backup_drives, backup_letter)
			}

			CreateSourceBackup(source_paths, backup_paths, archive_name)
			//CreateSourceBackup("C:/Users/tomas/Pictures/Backgrounds", "E:/backup", "test.7z")
			//fmt.Println(source, dest_drive_ksuid, dest_path)
			//insert_backups_db(source, dest_drive_ksuid, dest_drive_letter, dest_path)
			return nil
		},
	}

	removeSourceCmd = &cobra.Command{
		Use:   "remove-source -s [source id]",
		Short: "Remove source by source ids",
		RunE: func(cmd *cobra.Command, args []string) error {

			source_id, err := cmd.Flags().GetInt64("source-id")

			if err != nil {
				return err
			}

			if source_id == 0 {
				return errors.New("atleast one source id has to be passed")
			}

			RemoveSource(source_id)

			return nil
		},
	}

	loadDBFromDriveCmd = &cobra.Command{
		Use:   "load-db -d [drive letter]",
		Short: "Load database from drive",
		RunE: func(cmd *cobra.Command, args []string) error {

			drive_letter, err := cmd.Flags().GetString("drive-letter")

			if err != nil {
				return err
			}

			if (drive_letter < "a" || drive_letter > "z") && (drive_letter < "A" || drive_letter > "Z") {
				return errors.New("typed drive letter is not an alphabetic letter")
			}

			err = LoadDBFromDrive(drive_letter)

			if err != nil {
				return err
			}

			return nil
		},
	}

	removeDestinationCmd = &cobra.Command{
		Use:   "remove-backup -i [archive id] -d [drive ksuid] | -p [path to destination] | -a [archive name] | -l [drive letter]",
		Short: "Remove backup records by archive_id and drive_ksuid, path to destination, archive name or drive letter",
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

			drive_letter, err := cmd.Flags().GetString("drive-letter")

			if err != nil {
				return err
			}

			if drive_letter != "" {
				RemoveDestinationByDrive(drive_letter)
			}

			archive_name, err := cmd.Flags().GetString("archive-name")

			if err != nil {
				return err
			}

			if archive_name != "" {
				RemoveDestinationByArchive(archive_name)
			}

			return nil
		},
	}

	addSourceCmd = &cobra.Command{
		Use:   "add-source -s [source path] -a [archive name]",
		Short: "Add source to archive name",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(source_paths) == 0 {
				return errors.New("requires atleast one source path")
			}
			if archive_name == "" {
				return errors.New("requires archive name")
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

func TestRootCmd() *cobra.Command {
	return &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), args[0])
			return nil
		},
	}
}

func init() {
	//RemoveDestinationByArchive("crypto-images.7z")
	//RemoveSource(19)
	//RestoreFileDir(source_ids, []string{"test-archiv-epic-installer-a-obrazky.7z"}, backup_paths)
	//BackupFileDir([]int64{18, 19, 20})

	BackupDatabase()
	err := db.Fixtures()

	if err != nil {
		fmt.Println("Database tables couldnt be created.")
	}

	rootCmd.AddCommand(listDrivesCmd)
	rootCmd.AddCommand(listBackupsCmd)
	rootCmd.AddCommand(addDriveCmd)
	rootCmd.AddCommand(startBackupCmd)
	rootCmd.AddCommand(createBackupCmd)
	rootCmd.AddCommand(removeSourceCmd)
	rootCmd.AddCommand(removeDestinationCmd)
	rootCmd.AddCommand(startRestoreCmd)
	rootCmd.AddCommand(ClearAllTablesCmd)
	rootCmd.AddCommand(loadDBFromDriveCmd)

	addDriveCmd.Flags().StringP("drive-name", "n", "", "Drive name")

	removeSourceCmd.Flags().Int64P("source-id", "s", 0, "Source ID")

	removeDestinationCmd.Flags().Int64P("archive-id", "i", 0, "Archive ID")
	removeDestinationCmd.Flags().StringP("drive-ksuid", "d", "", "Drive Ksuid")
	removeDestinationCmd.Flags().StringP("dest-path", "p", "", "Destination path")
	removeDestinationCmd.Flags().StringP("archive-name", "a", "", "Archive name")
	removeDestinationCmd.Flags().StringP("drive-letter", "l", "", "Drive letter")

	//rootCmd.AddCommand(createBackupCmdTest)

	//createBackupCmd.Flags().StringSlice("source", source_paths, "sources paths")
	createBackupCmd.Flags().StringArrayVarP(&source_paths, "source", "s", make([]string, 0), "sources paths")
	createBackupCmd.Flags().StringArrayVarP(&backup_paths, "destination", "d", make([]string, 0), "destination path")
	createBackupCmd.Flags().StringVarP(&archive_name, "archive", "a", "", "archive name")
	createBackupCmd.MarkFlagRequired("source")
	createBackupCmd.MarkFlagRequired("destination")

	startBackupCmd.Flags().Int64SliceVarP(&source_ids, "source", "s", make([]int64, 0), "source ids")
	startBackupCmd.Flags().StringArrayVarP(&archive_names, "archive", "a", make([]string, 0), "archive name")

	//p *[]int64, name, shorthand string, value []int64, usage string
	startRestoreCmd.Flags().Int64SliceVarP(&source_ids, "source", "s", make([]int64, 0), "source ids")
	startRestoreCmd.Flags().StringArrayVarP(&archive_names, "archive", "a", make([]string, 0), "archive name")
	startRestoreCmd.Flags().StringArrayVarP(&backup_paths, "backup", "b", make([]string, 0), "backup paths")

	addSourceCmd.Flags().StringArrayVarP(&source_paths, "source", "s", make([]string, 0), "sources paths")
	addSourceCmd.Flags().StringVarP(&archive_name, "archive", "a", "", "archive name")

	loadDBFromDriveCmd.Flags().StringP("drive-letter", "l", "", "Drive letter")
	//RestoreFileDir(10)

	/*createBackupCmd.Flags().StringVarP(&source, "source", "s", "", "source path")
	createBackupCmd.Flags().StringVarP(&dest_drive_ksuid, "drive_ksuid", "k", "", "destination drive ksuid")
	createBackupCmd.Flags().StringVarP(&dest_drive_letter, "destination drive letter", "d", "", "destination drive letter")
	createBackupCmd.Flags().StringVarP(&dest_path, "additional path", "p", "", "additional path")*/

	//singleCmd.Flags().StringVarP(&options.File, "file", "f", "", "file containing urls. use - for stdin")
	//rootCmd.AddCommand(rootCmd)
}
