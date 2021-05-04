package cmd

import (
	"bufio"
	"database/sql"
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

//func AddSource()
func CreateSourceBackup(source_paths []string, backup_paths []string, archive_name string) {
	fmt.Println("Sources: " + strings.Join(source_paths, ", "))
	fmt.Println("Backups: " + strings.Join(backup_paths, ", "))

	//fmt.Println("-----------------------------------------------------------------------")

	for _, source_path := range source_paths {
		for _, backup_path := range backup_paths {

			//fmt.Println(source_path + " - " + backup_path)

			if helper.Exists(source_path) == nil && helper.Exists(backup_path) == nil {
				source_letter := strings.ReplaceAll(filepath.VolumeName(source_path), ":", "")
				backup_letter := strings.ReplaceAll(filepath.VolumeName(backup_path), ":", "")

				source_drive_ksuid := AddDrive(source_letter, "")
				backup_drive_ksuid := AddDrive(backup_letter, "")

				//fmt.Println("Source: " + source_letter + " - " + source_drive_ksuid)
				//fmt.Println("Backup: " + backup_letter + " - " + backup_drive_ksuid)

				if source_drive_ksuid != "" && backup_drive_ksuid != "" {
					source_path := helper.RemoveDriveLetter(source_path)
					backup_path := helper.RemoveDriveLetter(backup_path)

					//fmt.Println(source_path)
					//fmt.Println(backup_path)

					source_id := db.CreateSource(source_drive_ksuid, source_path)

					if source_id == 0 {
						continue
					}

					archive_ext := path.Ext(archive_name)

					if db.ArchiveExists(archive_name) {
						fmt.Println("Archive name isn't unique, so it was set to default with source id.")
						archive_name = ""
					}

					if archive_name == "" {
						archive_name = "backup-" + strconv.FormatInt(source_id, 10)
					}

					if archive_ext == "" {
						archive_name += ".7z"
					}

					archive_id := db.CreateArchive(archive_name)

					if archive_id == 0 {
						continue
					}

					res := db.CreateBackup(archive_id, backup_drive_ksuid, backup_path)

					if res == false {
						continue
					}

					archive_id = db.UpdateSourceArchive(source_id, archive_id)

					if archive_id == 0 {
						continue
					}

					// Vymazat vytvorene zaznamy pred continue ak sa vrati 0
				}

				//fmt.Println("Drives couldnt be added to DB.")

				//return 0
			}
		}
	}

	//fmt.Println("Files or directories dont exist.")

	//return 0
}

func ListBackups() {
	var backup_rels = db.FindBackups(0)
	var table_data [][]string
	var destinations []string

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
}

func TransformBackups(backup_rels map[int64]database.BackupRel) ([]string, string, string, []string) {
	var destinations []string
	var source string
	var archive_name string
	var backup_ksuids []string

	for _, element := range backup_rels {
		if len(source) == 0 {
			source = Ksuid2Drive(element.Source.Ksuid) + ":" + element.Source.Path.String
		}

		if len(archive_name) == 0 {
			archive_name = element.Archive.Name
		}

		for _, b := range element.Backup {
			destination_ksuid := Ksuid2Drive(b.Ksuid)
			if destination_ksuid != "" && b.Path.String != "" {
				destination := destination_ksuid + ":" + b.Path.String

				destinations = append(destinations, destination)
				backup_ksuids = append(backup_ksuids, b.Ksuid)

				err := os.MkdirAll(destination, os.ModePerm)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}

	return destinations, source, archive_name, backup_ksuids
}

func RestoreFileDir(source_id int64, backup_paths []string) {
	var backup_rels = db.FindBackups(source_id)

	destinations, source, archive_name, _ := TransformBackups(backup_rels)

	//fmt.Println(destinations)
	//fmt.Println(source)
	//fmt.Println(archive_name)

	_, err := os.Stat(source)

	if os.IsNotExist(err) {
		fmt.Println("File does not exist.")
	}

	fmt.Println("Restoring [" + strings.Join(destinations, "/"+archive_name+", ") + "] to " + source)

	cmd7zExists := helper.CommandAvailable("7z")
	path7z := "7z"

	if !cmd7zExists {
		_, err = os.Stat("7-ZipPortable/App/7-Zip64/7z.exe")

		if os.IsNotExist(err) {
			fmt.Println("7z executable isnt accesible.")
		}

		path7z = "7-ZipPortable/App/7-Zip64/7z.exe"
	}

	for _, destination := range destinations {
		archive_path := destination + "/" + archive_name
		_, err = os.Stat(archive_path)

		if len(backup_paths) > 0 && !helper.FindElm(backup_paths, archive_path) {
			continue
		}

		if os.IsNotExist(err) {
			fmt.Println("Couldnt be restored because archive doesn't exist.")
			continue
		}

		args := []string{"x", archive_path, "-y", "-o" + source}

		cmd := exec.Command(path7z, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println(err.Error())
			//return err
		}

		break
	}
}

func BackupFileDir(source_id int64) int {
	var backup_rels = db.FindBackups(source_id)

	destinations, source, archive_name, backup_ksuids := TransformBackups(backup_rels)

	_, err := os.Stat(source)

	if os.IsNotExist(err) {
		fmt.Println("Source file does not exist.")
		return 0
	}

	fmt.Println("Archiving " + source + " to [" + strings.Join(destinations, ", ") + "] " + archive_name)

	cmd7zExists := helper.CommandAvailable("7z")
	path7z := "7z"

	if !cmd7zExists {
		_, err = os.Stat("7-ZipPortable/App/7-Zip64/7z.exe")

		if os.IsNotExist(err) {
			fmt.Println("7z executable isnt accesible.")
			return 0
		}

		path7z = "7-ZipPortable/App/7-Zip64/7z.exe"
	}

	for _, destination := range destinations {
		var args []string
		archive_exists := helper.Exists(destination + "/" + archive_name)

		fmt.Println(archive_exists)

		if archive_exists == nil {
			args = []string{"u", destination + "/" + archive_name, source + "/*"}
		} else {
			args = []string{"a", "-t7z", destination + "/" + archive_name, source + "/*"}
		}

		cmd := exec.Command(path7z, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
	}

	for _, ksuid := range backup_ksuids {
		db.AddBackupTimestamp(source_id, ksuid)
	}

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
		fmt.Printf("Nepodarilo sa vytvorit subor %s\n", err)
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

func BackupDatabase() {
	drives := helper.GetDrives()
	var database_path = helper.GetDatabaseFile()

	old_timestamp := db.TestDatabase(database_path)

	//print(old_timestamp.Time.String())

	if len(drives) > 0 {
		for _, drive_letter := range drives {

			if helper.Exists(drive_letter+":/.drive") == nil {
				// ak ma .drive subor a nie je zapisane v db
				_, err := helper.ReadFileLines(drive_letter + ":/.drive")

				if err != nil {
					continue
				}

				drive_db_path := drive_letter + ":/sqlite-database.db"

				if helper.Exists(drive_db_path) != nil && helper.Exists(database_path) == nil {
					helper.CopyFile(database_path, drive_db_path)
					continue
				}

				new_timestamp := db.TestDatabase(drive_db_path)

				//print(new_timestamp.Time.String())

				if (new_timestamp == sql.NullTime{} && old_timestamp == sql.NullTime{}) {
					// Ziadne operacie
				}

				if new_timestamp.Time.After(old_timestamp.Time) {
					helper.CopyFile(drive_db_path, database_path)
					break
				}
				/*Pozriet na disky ci maju databazu ak hej tak precitat tabulku timestamps
				a porovnat ze ktore maju novsie zaznamy ak je lokalna novsia tak skopirovat */

			}
		}
	}

}

func NewDriveRecord(drive_letter string) database.DriveRecord {
	drive_record := database.DriveRecord{}
	drive_record.Letter = drive_letter
	drive_record.Name = ""
	drive_record.File_exists = false
	drive_record.File_accesible = false
	drive_record.Ksuid = ""
	drive_record.Database_exists = false
	drive_record.Database_inserted = false
	return drive_record
}

// Lists drives and their statuses
func ListDrives() {
	drives := helper.GetDrives()

	if len(drives) == 0 {
		fmt.Println("There are no drives connected to the PC.")
		return
	}

	var drive_records []database.DriveRecord

	for _, drive_letter := range drives {

		var drive_record = NewDriveRecord(drive_letter)

		if helper.Exists(drive_letter+":/.drive") != nil {
			fmt.Println(drive_letter + " - Drive doesn't have a .drive file")
			drive_records = append(drive_records, drive_record)
			continue
		}

		drive_record.File_exists = true

		// ak ma .drive subor a nie je zapisane v db
		lines, err := helper.ReadFileLines(drive_letter + ":/.drive")

		if err != nil {
			fmt.Println(drive_letter + " - .drive file isn't accesible.")
			drive_records = append(drive_records, drive_record)
			continue
		}

		drive_record.File_accesible = true
		drive_record.Ksuid = string(lines[0])

		drive_info, drive_name := db.DriveInDB(string(lines[0]))

		// If .drive exists but isnt in db
		//fmt.Print(" - " + string(lines[0]))
		if drive_info != "" {
			drive_record.Name = drive_name
			drive_records = append(drive_records, drive_record)
			continue
		}

		drive_record.Database_exists = true

		//fmt.Println(" - Drive has .drive file but werent in drives table.")

		id := db.InsertDriveDB(string(lines[0]), "")

		if id <= 0 {
			fmt.Println("Inserting " + drive_letter + " drive record into DB failed.")
			drive_records = append(drive_records, drive_record)
			continue
		}

		fmt.Println(drive_letter + " drive was successfully inserted into DB.")

		drive_record.Database_inserted = true

		//fmt.Println("Drive was added to DB.")
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Drive Letter", "Custom Name", ".drive exists", ".drive accesible", "Ksuid", "In DB", "Inserted into DB"})

	for _, v := range drive_records {
		var table_row []string
		table_row = append(table_row, v.Letter)
		table_row = append(table_row, v.Name)
		table_row = append(table_row, strconv.FormatBool(v.File_exists))
		table_row = append(table_row, strconv.FormatBool(v.File_accesible))
		table_row = append(table_row, v.Ksuid)
		table_row = append(table_row, strconv.FormatBool(v.Database_exists))
		table_row = append(table_row, strconv.FormatBool(v.Database_inserted))

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
	if helper.Exists(drive_letter+":/.drive") != nil {
		ksuid := helper.GenKsuid()

		drive_info, _ := db.DriveInDB(ksuid)

		if drive_info == "" {
			id := db.InsertDriveDB(ksuid, drive_name)

			if id > 0 {
				success := CreateDiskIdentityFile(drive_letter, ksuid)
				if success {
					return ksuid
				}
			}

			return ""
		}
	}

	return helper.GetKsuidFromDrive(drive_letter)
}

func RemoveSource(source_id int64) int {
	res := db.RemoveSource(source_id)

	if res {
		fmt.Println("Destination was removed succesfuly.")
		return 1
	}

	fmt.Println("Destination was not removed.")
	return 0
}

func RemoveDestination(archive_id int64, drive_ksuid string) int {
	res := db.RemoveDestination(archive_id, drive_ksuid)

	if res {
		fmt.Println("Destination was removed succesfuly.")
		return 1
	}

	fmt.Println("Destination was not removed.")
	return 0
}

func RemoveDestinationByPath(destination_path string) {
	dest_letter := strings.ReplaceAll(filepath.VolumeName(destination_path), ":", "")
	dest_drive_ksuid := AddDrive(dest_letter, "")

	if len(dest_drive_ksuid) > 0 {
		dest_archive_name := filepath.Base(destination_path)
		if len(dest_archive_name) > 0 {
			db.RemoveDestinationByPath(dest_archive_name, dest_drive_ksuid)
		}
	}
}
