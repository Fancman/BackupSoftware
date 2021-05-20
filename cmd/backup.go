package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Fancman/BackupSoftware/database"
	helper "github.com/Fancman/BackupSoftware/helpers"
	_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
)

func AddSource(source_paths []string, archive_name string) {
	for _, source_path := range source_paths {
		if helper.Exists(source_path) == nil {

			source_letter := strings.ReplaceAll(filepath.VolumeName(source_path), ":", "")
			source_drive_ksuid := AddDrive(source_letter, "")

			if source_drive_ksuid != "" {
				source_path := helper.RemoveDriveLetter(source_path)
				source_id := db.CreateSource(source_drive_ksuid, source_path)

				if source_id == 0 || source_id == -1 {
					continue
				}

				archive_id := db.GetArchiveID(archive_name)

				if archive_id == 0 {
					continue
				}

				db.UpdateSourceArchive(source_id, archive_id)
			}
		}

	}
}

// Creates records for source-backup relation
// Archive name is optional
func CreateSourceBackup(source_paths []string, backup_paths []string, archive_name string) error {
	fmt.Println("Sources: " + strings.Join(source_paths, ", "))
	fmt.Println("Backups: " + strings.Join(backup_paths, ", "))

	archive_ext := path.Ext(archive_name)
	archive_name = helper.FileNameWithoutExtension(archive_name)
	default_archive_name := "backup-" + strconv.FormatInt(time.Now().Unix(), 10)

	if archive_name == "" {
		archive_name = default_archive_name
	}

	if archive_ext == "" {
		archive_ext = ".7z"
	}

	archive_id, err := db.CreateArchive(archive_name + archive_ext)

	if err != nil {
		fmt.Println(err)

		archive_id, err = db.CreateArchive(default_archive_name + archive_ext)

		fmt.Println("Archive name was set to: " + default_archive_name + archive_ext)

		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	for _, source_path := range source_paths {
		if helper.Exists(source_path) == nil {

			source_letter := strings.ReplaceAll(filepath.VolumeName(source_path), ":", "")
			source_drive_ksuid := AddDrive(source_letter, "")

			if source_drive_ksuid != "" {
				source_path := helper.RemoveDriveLetter(source_path)
				source_id := db.CreateSource(source_drive_ksuid, source_path)

				if source_id == 0 {
					continue
				}

				if source_id == -1 {
					continue
				}

				err = db.UpdateSourceArchive(source_id, archive_id)

				if err == nil {
					continue
				}
				// Vymazat vytvorene zaznamy pred continue ak sa vrati 0
			}
		}

	}

	for _, backup_path := range backup_paths {

		if helper.Exists(backup_path) == nil {
			backup_letter := strings.ReplaceAll(filepath.VolumeName(backup_path), ":", "")

			backup_drive_ksuid := AddDrive(backup_letter, "")

			if backup_drive_ksuid != "" {
				backup_path := helper.RemoveDriveLetter(backup_path)

				err := db.CreateBackup(archive_id, backup_drive_ksuid, backup_path)

				if err != nil {
					continue
				}
			}
		}
	}

	return nil
}

func ListBackups() int {
	backup_rels, err := db.FindBackups([]int64{}, []string{})
	var table_data [][]string
	var destinations []string

	if err != nil {
		return 0
	}

	for key, element := range backup_rels {
		var table_row []string

		//fmt.Print("Source id: " + strconv.FormatInt(key, 10))

		table_row = append(table_row, strconv.FormatInt(key, 10))

		drive_letter := Ksuid2Drive(element.Source.Ksuid)

		if drive_letter == "" {
			table_row = append(table_row, "Not accesible")
			//fmt.Print(" - Source drive isn't accesible. " + " [" + element.Source.Path.String + "]")
		} else {
			table_row = append(table_row, drive_letter)
			//fmt.Print(" - [" + drive_letter + ":" + element.Source.Path.String + "]")
		}

		table_row = append(table_row, element.Source.Path.String)

		//fmt.Print(" [")
		for _, b := range element.Backup {
			//fmt.Println(Ksuid2Drive(b.Ksuid))
			destination_ksuid := Ksuid2Drive(b.Ksuid)
			if destination_ksuid != "" && b.Path.String != "" {
				destinations = append(destinations, destination_ksuid+":"+b.Path.String)
				//fmt.Print(Ksuid2Drive(b.Ksuid) + ":" + b.Path.String)
			}
		}
		//fmt.Print("]")

		table_row = append(table_row, strings.Join(destinations, ", "))

		destinations = nil

		table_row = append(table_row, element.Archive.Name)

		//fmt.Print(" - Archive name: " + element.Archive.Name)
		if element.Archived_at.Valid {
			table_row = append(table_row, element.Archived_at.Time.Local().Format(time.UnixDate))
			//fmt.Print(" - Archived at : " + element.Archived_at.Time.Local().Format(time.UnixDate))
		} else {
			table_row = append(table_row, "Nil")
		}
		//fmt.Print("\n")
		table_data = append(table_data, table_row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Source ID", "Source drive", "Source Path", "Destinations", "Archive", "Archived at"})

	for _, v := range table_data {
		table.Append(v)
	}
	table.Render()

	return 1
}

func TransformBackups(backup_rels map[int64]database.BackupRel) map[string]database.BackupPaths {
	var backup_paths = make(map[string]database.BackupPaths)

	for _, element := range backup_rels {

		_, ok := backup_paths[element.Archive.Name]

		// backup_paths doesnt have key with archive name
		source_path := Ksuid2Drive(element.Source.Ksuid) + ":" + element.Source.Path.String

		for _, b := range element.Backup {
			destination_ksuid := Ksuid2Drive(b.Ksuid)
			if destination_ksuid != "" && b.Path.String != "" {
				destination_path := destination_ksuid + ":" + b.Path.String

				if !ok {
					backup_paths[element.Archive.Name] = database.BackupPaths{
						Sources:     []string{source_path},
						SourceIDs:   []int64{element.Source.Id},
						Destination: destination_path,
						BackupKsuid: b.Ksuid,
					}

					continue
				}

				var backup_path = backup_paths[element.Archive.Name]
				var sources []string = backup_paths[element.Archive.Name].Sources
				sources = append(sources, source_path)
				var sources_ids []int64 = backup_paths[element.Archive.Name].SourceIDs
				sources_ids = append(sources_ids, element.Source.Id)
				backup_path.Sources = sources
				backup_path.SourceIDs = sources_ids
				backup_paths[element.Archive.Name] = backup_path
			}
		}
	}

	return backup_paths

	//return destinations, source, archive_name, backup_ksuids
}

func RestoreFileDir(source_ids []int64, archive_names []string, backup_paths []string) int {
	backup_rels, err := db.FindBackups(source_ids, archive_names)

	if err != nil {
		return 0
	}

	backup_paths_rel := TransformBackups(backup_rels)

	for archive_name, backup_path := range backup_paths_rel {
		archive_path := backup_path.Destination + "/" + archive_name

		_, err := os.Stat(archive_path)

		if os.IsNotExist(err) {
			fmt.Println("Archive file does not exist.")
		}

		cmd7zExists := helper.CommandAvailable("7z")
		path7z := "7z"

		if !cmd7zExists {
			_, err = os.Stat("7-ZipPortable/App/7-Zip64/7z.exe")

			if os.IsNotExist(err) {
				fmt.Println("7z executable is not accesible.")
			}

			path7z = "7-ZipPortable/App/7-Zip64/7z.exe"
		}

		for _, source_path := range backup_path.Sources {
			var removed bool = false
			var output_path string = ""

			source_parts := strings.Split(source_path, `\`)

			if len(source_parts) > 1 {
				for i := len(source_parts) - 1; i >= 0; i-- {
					if source_parts[i] != "" {
						if !removed {
							removed = true
						} else {
							output_path = source_parts[i] + `\` + output_path
						}
					}
				}
			}

			if len(output_path) == 0 {
				output_path = `\`
			}

			args := []string{"x", archive_path, "-y", "-o" + output_path, filepath.Base(source_path)}

			cmd := exec.Command(path7z, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				fmt.Println(err.Error())
				//return err
			}

		}
	}
	return 1
}

func BackupFileDir(source_ids []int64, archive_names []string) int {
	backup_rels, err := db.FindBackups(source_ids, archive_names)

	if err != nil {
		return 0
	}

	backup_paths := TransformBackups(backup_rels)

	for archive_name, backup_path := range backup_paths {
		for _, source := range backup_path.Sources {
			_, err := os.Stat(source)

			if os.IsNotExist(err) {
				fmt.Println("Source file or directory do not exist.")
				return 0
			}
		}

		fmt.Println("Archiving [" + strings.Join(backup_path.Sources, ", ") + "] to " + backup_path.Destination + `\` + archive_name)

		cmd7zExists := helper.CommandAvailable("7z")
		path7z := "7z"

		if !cmd7zExists {
			_, err := os.Stat("7-ZipPortable/App/7-Zip64/7z.exe")

			if os.IsNotExist(err) {
				fmt.Println("7z executable isnt accesible.")
				return 0
			}

			path7z = "7-ZipPortable/App/7-Zip64/7z.exe"
		}

		var args []string
		archive_exists := helper.Exists(backup_path.Destination + "/" + archive_name)

		fmt.Println(archive_exists)

		if archive_exists == nil {
			args = []string{"u", backup_path.Destination + "/" + archive_name}
		} else {
			args = []string{"a", "-t7z", backup_path.Destination + "/" + archive_name}
		}

		args = append(args, backup_path.Sources...)

		cmd := exec.Command(path7z, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()

		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		drive_letter := Ksuid2Drive(backup_path.BackupKsuid)

		// Add timestamp records
		for _, source_id := range backup_path.SourceIDs {
			db.AddBackupTimestamp(source_id, backup_path.BackupKsuid)
		}

		// Copy local database file on drive which we used as backup
		if drive_letter != "" {
			SpreadDatabase(drive_letter)
		}

	}

	/*for _, ksuid := range backup_ksuids {
		drive_letter := Ksuid2Drive(ksuid)
		db.AddBackupTimestamp(source_id, ksuid)
		// Copy local database file on drive which we used as backup
		if drive_letter != "" {
			SpreadDatabase(drive_letter)
		}
	}*/

	return 1
}

// Creates .drive file with ksuid in it
func CreateDiskIdentityFile(drive_letter string, ksuid string) bool {
	currentTime := time.Now()

	data := []string{
		ksuid,
		currentTime.String(),
	}

	//file, err := os.OpenFile(drive_letter+":/.drive", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	file, err := os.Create(drive_letter + ":/.drive")

	if err != nil {
		fmt.Printf("Drive does not exist or is not accessible. %s\n", err)
		db.DelDriveDB(ksuid)
		return false
	}

	defer file.Close()

	datawriter := bufio.NewWriter(file)

	for _, data := range data {
		_, _ = datawriter.WriteString(data + "\n")
	}

	datawriter.Flush()
	file.Close()

	return true
}

func SpreadDatabase(drive_letter string) int {
	var database_path = helper.GetDatabaseFile()
	var drive_db_path = drive_letter + ":/sqlite-database.db"
	var ksuid = helper.GetKsuidFromDrive(drive_letter)

	if ksuid == "" {
		return 0
	}

	if helper.Exists(database_path) == nil {
		helper.CopyFile(database_path, drive_db_path)
	}

	return 1
}

func LoadDBFromDrive(drive_letter string) error {
	var database_path = helper.GetDatabaseFile()
	var source_path = drive_letter + ":/sqlite-database.db"
	err := helper.Exists(source_path)

	if err != nil {
		return err
	}

	err = helper.CopyFile(source_path, database_path)

	if err != nil {
		return err
	}

	fmt.Println("New database was copied on place where local database should be")

	return nil
}

func GetDrivesWithDB() []string {
	var drives = helper.GetDrives()
	var drives_db = []string{}

	if len(drives) > 0 {
		for _, drive_letter := range drives {

			if helper.Exists(drive_letter+":/.drive") == nil {
				ksuid := helper.GetKsuidFromDrive(drive_letter)

				if ksuid == "" {
					continue
				}

				drive_db_path := drive_letter + ":/sqlite-database.db"

				if helper.Exists(drive_db_path) == nil {
					drives_db = append(drives_db, drive_letter)
				}

			}
		}
	}

	return drives_db
}

func BackupDatabase() int {
	var database_path = helper.GetDatabaseFile()
	var drives = helper.GetDrives()
	var timestamp_map = map[string]map[int64]database.Timestamp{}

	// Iterate available drives and if they have
	// .drive file and not database file then copy it onto them
	// If Local database does not exist collect records from timestamp
	// tables located on drives.
	if len(drives) > 0 {
		for _, drive_letter := range drives {

			if helper.Exists(drive_letter+":/.drive") == nil {
				ksuid := helper.GetKsuidFromDrive(drive_letter)

				if ksuid == "" {
					continue
				}

				drive_db_path := drive_letter + ":/sqlite-database.db"

				if helper.Exists(drive_db_path) != nil && helper.Exists(database_path) == nil {
					SpreadDatabase(drive_letter)
					continue
				}

				if helper.Exists(drive_db_path) == nil && helper.Exists(database_path) != nil {
					timestamp_map[drive_letter] = db.TestDatabase(drive_db_path)
				}

			}
		}
	}

	var timestamp_records = db.TestDatabase(database_path)

	// Local timestamp table has no records &&
	//  other drives have records in timestamp table
	if len(timestamp_records) == 0 && len(timestamp_map) > 0 {
		var newest_timestamp_records = map[int64]database.Timestamp{}
		var newest_drive_letter = ""

		// iterate throught drives timestamp tables
		for drive_letter, test_timestamp_records := range timestamp_map {
			// set first drive timestamp records as newest
			if newest_drive_letter == "" {
				newest_drive_letter = drive_letter
				newest_timestamp_records = test_timestamp_records
				continue
			}
			// comparing archived_at timestamps
			for source_id, timestamp := range test_timestamp_records {
				var timestamp_row, ok = newest_timestamp_records[source_id]
				if ok {
					if timestamp.Archived_at.Time.After(timestamp_row.Archived_at.Time) {
						newest_timestamp_records = test_timestamp_records
						newest_drive_letter = drive_letter
					}
				}
			}

		}

		// Copy database file from drive with newest timestamp records.
		if newest_drive_letter != "" {
			helper.CopyFile(newest_drive_letter+":/sqlite-database.db", database_path)
			return 1
		}

	}

	// Create new database file if there wasnt any on other drives
	if len(timestamp_records) == 0 && len(timestamp_map) == 0 {
		err := database.CreateDB()

		fmt.Println("CREATED NEW DB FILE")

		if err != nil {
			fmt.Println(err)
		}
	}

	return 1
}

func NewDriveRecord(drive_letter string) database.DriveRecord {
	drive_record := database.DriveRecord{}
	drive_record.Letter = drive_letter
	drive_record.Name = ""
	drive_record.File_exists = false
	drive_record.File_accesible = false
	drive_record.Ksuid = ""
	drive_record.Timestamp = ""
	return drive_record
}

// Lists drives and their statuses
func ListDrives() {
	var database_path = helper.GetDatabaseFile()
	var base_db_volume = strings.ReplaceAll(filepath.VolumeName(database_path), ":", "")
	drives := helper.GetDrives()

	if len(drives) == 0 {
		fmt.Println("There are no drives connected to the PC.")
		return
	}

	var drive_records []database.DriveRecord

	for _, drive_letter := range drives {

		var drive_record = NewDriveRecord(drive_letter)

		if helper.Exists(drive_letter+":/.drive") != nil {
			fmt.Println(drive_letter + " - Drive does not have a .drive file")
			drive_records = append(drive_records, drive_record)
			continue
		}

		if drive_letter == base_db_volume {
			if db.GetNewestTimestamp(database_path).Valid {
				drive_record.Timestamp = db.GetNewestTimestamp(database_path).Time.String()
			}
		}

		if drive_record.Timestamp == "" {
			var drive_db_path = drive_letter + ":/sqlite-database.db"

			if db.GetNewestTimestamp(drive_db_path).Valid {
				drive_record.Timestamp = db.GetNewestTimestamp(drive_db_path).Time.String()
			}
		}

		drive_record.File_exists = true
		drive_record.Ksuid = helper.GetKsuidFromDrive(drive_letter)

		if drive_record.Ksuid == "" {
			fmt.Println(drive_letter + " - .drive file is not accessible.")
			drive_records = append(drive_records, drive_record)
			continue
		}

		drive_record.File_accesible = true

		drive_info, drive_name := db.DriveInDB(drive_record.Ksuid)

		if drive_info != "" {
			drive_record.Name = drive_name.String
			drive_records = append(drive_records, drive_record)
			continue
		}

		res := db.InsertDriveDB(drive_record.Ksuid, "")

		if res <= 0 {
			fmt.Println("Inserting " + drive_letter + " drive record into DB failed.")
			drive_records = append(drive_records, drive_record)
			continue
		}

		fmt.Println(drive_letter + " drive was successfully inserted into DB.")

		drive_records = append(drive_records, drive_record)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Drive Letter", "Custom Name", ".drive exists", ".drive accesible", "Ksuid", "Latest Timestamp"})

	for _, d := range drive_records {
		var table_row []string
		table_row = append(table_row, d.Letter)
		table_row = append(table_row, d.Name)
		table_row = append(table_row, strconv.FormatBool(d.File_exists))
		table_row = append(table_row, strconv.FormatBool(d.File_accesible))
		table_row = append(table_row, d.Ksuid)
		table_row = append(table_row, d.Timestamp)

		table.Append(table_row)
	}
	table.Render()
}

/*func DriveLetter2Ksuid(drive_letter string) (string, error) {
	err := helper.DriveExists(drive_letter)
	if err == nil {
		if helper.Exists(drive_letter + ":/.drive") {

		}
	} else {
		return drive_letter, err
	}

	return drive_letter, nil
}*/

// Get path to drive by ksuid
func Ksuid2Drive(ksuid string) string {
	drives := helper.GetDrives()

	if len(drives) > 0 {
		for _, drive_letter := range drives {

			err := helper.Exists(drive_letter + ":/.drive")

			if err == nil {
				// ak ma .drive subor a nie je zapisane v db
				lines, err := helper.ReadFileLines(drive_letter + ":/.drive")

				if err != nil {
					fmt.Printf("Error while reading a file: %s", err)
				}

				if ksuid == lines[0] {
					return drive_letter
				}
			}

		}
	}

	return ""
}

func AddDrive(drive_letter string, drive_name string) string {
	var ksuid string

	// Drive doesnt have .drive file
	if helper.Exists(drive_letter+":/.drive") != nil {
		ksuid = helper.GenKsuid()

		// is generated ksuid in DB?
		drive_info, _ := db.DriveInDB(ksuid)

		// Ksuid from .drive isnt in DB
		if drive_info == "" {
			id := db.InsertDriveDB(ksuid, drive_name)

			if id > 0 {
				fmt.Println("Drive was succesfully inserted into DB.")
				success := CreateDiskIdentityFile(drive_letter, ksuid)
				if success {
					fmt.Println(".drive file was succesfully created.")
					return ksuid
				}
			}

			return ""
		}

		AddDrive(drive_letter, drive_name)
	}

	ksuid = helper.GetKsuidFromDrive(drive_letter)

	if ksuid == "" {
		fmt.Println("Drive don't have .drive file accessible.")
		return ""
	}

	drive_info, saved_drive_name := db.DriveInDB(ksuid)

	// Update drive name
	if drive_name != "" && drive_name != saved_drive_name.String {
		res := db.UpdateDriveName(ksuid, drive_name)
		if res == 1 {
			fmt.Println("Drive name was updated.")
		} else {
			fmt.Println("Drive name cant be updated.")
		}
	}

	if drive_info != "" {
		fmt.Println("Drive was already in DB and .drive file is created.")
		return ksuid
	}
	// Drive isnt in db
	res := db.InsertDriveDB(ksuid, drive_name)

	if res <= 0 {
		fmt.Println("Drive cant be inserted into DB.")
		return ""
	}

	fmt.Println("Drive was succesfully inserted into DB.")

	return ksuid
}

func RemoveSource(source_id int64) error {
	archive_id := db.GetSourceArchive(source_id)

	if archive_id > 0 {
		RemoveUnusedArchive(archive_id, "source")
	}

	err := db.RemoveSource(source_id)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func RemoveUnusedArchive(archive_id int64, archive_usage string) bool {
	source_occur, backup_occur := db.ArchiveUsed(archive_id)

	if archive_usage == "source" {
		source_occur -= 1
	}

	if archive_usage == "destination" {
		backup_occur -= 1
	}

	if source_occur == 0 || backup_occur == 0 {
		if source_occur > 0 {
			err := db.RemoveSources(archive_id)
			if err {
				fmt.Println("Sources without backup records were deleted.")
			}
		}

		if backup_occur > 0 {
			err := db.RemoveDestinations(archive_id)
			if err {
				fmt.Println("Backups without source records were deleted.")
			}
		}

		res := db.DelArchiveDB(archive_id)

		if res {
			fmt.Println("Archive was deleted")
			return true
		}

		fmt.Println("Archive couldnt be deleted.")
		return false
	}

	fmt.Println("Archive couldnt be deleted because it is used in " + strconv.Itoa(source_occur+backup_occur-1) + " more records.")

	return false
}

func RemoveDestinationByDrive(drive_letter string) int {
	drive_ksuid := helper.GetKsuidFromDrive(drive_letter)

	if drive_ksuid != "" {
		res := db.RemoveDestinationByDrive(drive_ksuid)
		if res {
			source_ids := db.GetSourcesWithoutBackup()
			archive_ids := db.GetArchivesWithoutBackup()

			for _, source_id := range source_ids {
				db.RemoveSource(source_id)
				fmt.Println("Unused source was deleted with id: " + strconv.FormatInt(source_id, 10))
			}

			for _, archive_id := range archive_ids {
				db.DelArchiveDB(archive_id)
				fmt.Println("Unused archive was deleted with id: " + strconv.FormatInt(archive_id, 10))
			}

			return 1
		}
	}

	return 0
}

func RemoveDestinationByArchive(archive_name string) int {
	archive_id := db.GetArchiveID(archive_name)

	if archive_id > 0 {
		res := db.RemoveDestinationByArchive(archive_id)
		if res {
			source_ids := db.GetSourcesWithoutBackup()
			archive_ids := db.GetArchivesWithoutBackup()

			for _, source_id := range source_ids {
				db.RemoveSource(source_id)
				fmt.Println("Unused source was deleted with id: " + strconv.FormatInt(source_id, 10))
			}

			fmt.Println(strconv.Itoa(len(source_ids)) + " unused sources were deleted.")

			for _, archive_id := range archive_ids {
				db.DelArchiveDB(archive_id)
				fmt.Println("Unused archive was deleted with id: " + strconv.FormatInt(archive_id, 10))
			}

			return 1
		}
	}

	return 0
}

func RemoveDestination(archive_id int64, drive_ksuid string) int {
	res := db.RemoveDestination(archive_id, drive_ksuid)

	if res {
		RemoveUnusedArchive(archive_id, "destination")
		fmt.Println("Destination was removed successfully.")
		return 1
	}

	fmt.Println("Destination was not removed.")
	return 0
}

func RemoveDestinationByPath(destination_path string) {
	destination_path = helper.CleanPath(destination_path)

	dest_letter := strings.ReplaceAll(filepath.VolumeName(destination_path), ":", "")
	dest_drive_ksuid := AddDrive(dest_letter, "")

	if len(dest_drive_ksuid) > 0 {
		dest_archive_name := filepath.Base(destination_path)
		if len(dest_archive_name) > 0 {
			res := db.RemoveDestinationByPath(dest_archive_name, dest_drive_ksuid)

			if res {
				archive_id := db.GetArchiveID(archive_name)
				if archive_id > 0 {
					RemoveUnusedArchive(archive_id, "destination")
				}
			}
		}
	}
}
