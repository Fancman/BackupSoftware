package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Fancman/BackupSoftware/database"
	helper "github.com/Fancman/BackupSoftware/helpers"
	_ "github.com/mattn/go-sqlite3"
)

//var db *sql.DB
var err error

type Destination struct {
	ksuid string
	path  string
}

type Backup struct {
	id           int
	source       string
	destinations []Destination
}

type Archive struct {
	id   int
	name string
}

type Drive struct {
	ksuid string
	name  string
}

type Source struct {
	id      int
	archive Archive
	path    string
}

//func AddSource()
func CreateSourceBackup(source_path string, backup_path string, archive_name string) {
	if helper.Exists(source_path) && helper.Exists(backup_path) {
		source_letter := strings.ReplaceAll(filepath.VolumeName(source_path), ":", "")
		backup_letter := strings.ReplaceAll(filepath.VolumeName(backup_path), ":", "")

		source_drive_ksuid := AddDrive(source_letter)
		backup_drive_ksuid := AddDrive(backup_letter)

		fmt.Println("Source: " + source_letter + " - " + source_drive_ksuid)
		fmt.Println("Backup: " + backup_letter + " - " + backup_drive_ksuid)

		if source_drive_ksuid != "" && backup_drive_ksuid != "" {
			source_path := helper.RemoveDriveLetter(source_path)
			backup_path := helper.RemoveDriveLetter(backup_path)

			//fmt.Println(source_path)
			//fmt.Println(backup_path)

			source_id := db.CreateSource(source_drive_ksuid, source_path)

			//fmt.Println(source_id)

			if archive_name == "" {
				archive_name = "backup-" + strconv.FormatInt(source_id, 10)
			}

			archive_id := db.CreateArchive(archive_name)

			db.CreateBackup(archive_id, backup_drive_ksuid, backup_path)
			db.UpdateSourceArchive(source_id, archive_id)
		} else {
			fmt.Println("Drives couldnt be added to DB.")
		}
	} else {
		fmt.Println("Files or directories dont exist.")
		//fmt.Println(helper.Exists(source_path))
		//fmt.Println(helper.Exists(backup_path))
	}
}

func ListBackups() {
	var backup_rels = db.FindBackups(0)

	for key, element := range backup_rels {
		fmt.Print("Source id: " + strconv.FormatInt(key, 10))

		fmt.Print(" - " + Ksuid2Drive(element.Source.Ksuid) + ":" + element.Source.Path.String)

		fmt.Print(" [")
		for _, b := range element.Backup {
			//fmt.Println(Ksuid2Drive(b.Ksuid))
			if Ksuid2Drive(b.Ksuid) != "" && b.Path.String != "" {
				fmt.Print(Ksuid2Drive(b.Ksuid) + ":" + b.Path.String)
			}
		}
		fmt.Print("]")

		fmt.Print(" - Archive name: " + element.Archive.Name)
		fmt.Print("\n")
	}
}

func TransformBackups(backup_rels map[int64]database.BackupRel) ([]string, string, string) {
	var destinations []string
	var source string
	var archive_name string

	for _, element := range backup_rels {
		if len(source) == 0 {
			source = Ksuid2Drive(element.Source.Ksuid) + ":" + element.Source.Path.String
		}

		if len(archive_name) == 0 {
			archive_name = element.Archive.Name
		}

		for _, b := range element.Backup {
			if Ksuid2Drive(b.Ksuid) != "" && b.Path.String != "" {
				destination := Ksuid2Drive(b.Ksuid) + ":" + b.Path.String
				destinations = append(destinations, destination)

				err := os.MkdirAll(destination, os.ModePerm)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}

	return destinations, source, archive_name
}

func BackupFileDir(source_id int64) {
	var backup_rels = db.FindBackups(source_id)

	destinations, source, archive_name := TransformBackups(backup_rels)

	info, err := os.Stat(source)

	if os.IsNotExist(err) {
		fmt.Println("File does not exist.")
	}

	if info.IsDir() {
		fmt.Println("temp is a directory")
	} else {
		fmt.Println("temp is a file")
	}

	fmt.Println("Archiving " + source + " to [" + strings.Join(destinations, ", ") + "] " + archive_name)

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
		args := []string{"a", "-t7z", destination + "/" + archive_name, source + "/*"}

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

// Returns records from table 'backups'
func find_backup(db *sql.DB, backup_id int) Backup {
	var id int
	var source string
	var destination_ksuid string
	var path string
	//var cron string
	stmt := `SELECT b.id, b.source, dr.drive_ksuid, de.path FROM backups b JOIN destinations de ON de.backup_id = b.id JOIN drives dr ON dr.drive_ksuid=de.drive_ksuid WHERE b.id = ?;`

	rows, err := db.Query(stmt, backup_id)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var backup Backup
	var i = 0

	for rows.Next() {
		err := rows.Scan(&id, &source, &destination_ksuid, &path)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(id, source, destination_ksuid)
		destionation := Destination{ksuid: destination_ksuid, path: path}

		if i == 0 {
			backup = Backup{id: id, source: source, destinations: []Destination{destionation}}
		} else {
			backup.destinations = append(backup.destinations, destionation)
		}

		i++
	}

	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	return backup
}

// Returns records from table 'backups'
func find_backups(db *sql.DB, backup_id int) []Backup {
	var id int
	var source string
	var destination_ksuid string
	var path string
	//var cron string
	stmt := `SELECT b.id, b.source, dr.drive_ksuid, de.path FROM backups b JOIN destinations de ON de.backup_id = b.id JOIN drives dr ON dr.drive_ksuid=de.drive_ksuid WHERE b.id = ?;`

	rows, err := db.Query(stmt, backup_id)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var backups []Backup

	for rows.Next() {
		var exists_in_backups = false

		err := rows.Scan(&id, &source, &destination_ksuid, &path)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(id, source, destination_ksuid)

		for k, v := range backups {
			if v.id == id {
				backups[k].destinations = append(backups[k].destinations, Destination{ksuid: destination_ksuid, path: path})
				exists_in_backups = true
				break
			}
		}
		if !exists_in_backups {
			destionation := Destination{ksuid: destination_ksuid, path: path}
			backups = append(backups, Backup{id: id, source: source, destinations: []Destination{destionation}})
		}
	}

	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	return backups
}

// Returns records from table 'backups'
func list_backups(db *sql.DB) []Backup {
	var id int
	var source string
	var destination_ksuid string
	var path string
	//var cron string
	stmt := `SELECT b.id, b.source, dr.drive_ksuid, de.path FROM backups b JOIN destinations de ON de.backup_id = b.id JOIN drives dr ON dr.drive_ksuid=de.drive_ksuid`

	rows, err := db.Query(stmt)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var backups []Backup

	for rows.Next() {
		var exists_in_backups = false

		err := rows.Scan(&id, &source, &destination_ksuid, &path)
		if err != nil {
			fmt.Println(err)
		}

		//fmt.Println(id, source, destination_ksuid)

		for k, v := range backups {
			if v.id == id {
				backups[k].destinations = append(backups[k].destinations, Destination{ksuid: destination_ksuid, path: path})
				exists_in_backups = true
				break
			}
		}
		if !exists_in_backups {
			destionation := Destination{ksuid: destination_ksuid, path: path}
			backups = append(backups, Backup{id: id, source: source, destinations: []Destination{destionation}})
		}
	}

	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	return backups
}

// Returns drive by letter from db
/*func drive_db_exists_letter(drive_letter string) (bool, []string) {
	stmt := `SELECT drive_letter, drive_ksuid FROM drives
	WHERE drive_letter=?`
	var r_ksuid string
	var r_letter string
	row := db.QueryRow(stmt, drive_letter)
	switch err := row.Scan(&r_letter, &r_ksuid); err {
	case sql.ErrNoRows:
		return false, []string{}
	case nil:
		return true, []string{r_letter, r_ksuid}
	default:
		panic(err)
	}
}*/

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

// Lists drives and their statuses
func ListDrives() {
	drives := helper.GetDrives()

	if len(drives) > 0 {
		for _, drive_letter := range drives {

			fmt.Print(drive_letter)

			if helper.Exists(drive_letter + ":/.drive") {
				// ak ma .drive subor a nie je zapisane v db
				lines, err := helper.ReadFileLines(drive_letter + ":/.drive")

				if err != nil {
					fmt.Printf("Error pri citani suboru: %s", err)
				}

				drive_info := db.DriveInDB(string(lines[0]))

				// If .drive exists but isnt in db
				fmt.Print(" - " + string(lines[0]))
				if drive_info == "" {
					id := db.InsertDriveDB(string(lines[0]))

					fmt.Print(" - Drive has .drive file but werent in drives table.")

					if id > 0 {
						fmt.Println("Drive was added to DB.")
					}

					fmt.Println("Inserting drive record into DB failed.")
				} else {
					fmt.Print(" - Drive has .drive file and is in drives table.")
				}
			} else {
				fmt.Print(" - Drive doesn't have a .drive file")
			}

			fmt.Print("\n")
		}
	} else {
		fmt.Println("There are none drives connected to PC.")
	}
}

func DriveLetter2Ksuid(drive_letter string) (string, error) {
	err := helper.DriveExists(drive_letter)
	if err == nil {
		if helper.Exists(drive_letter + ":/.drive") {

		}
	} else {
		return drive_letter, err
	}

	return drive_letter, nil
}

// Get path to drive by ksuid
func Ksuid2Drive(ksuid string) string {
	drives := helper.GetDrives()

	if len(drives) > 0 {
		for _, drive_letter := range drives {

			//fmt.Print(drive_letter)

			if helper.Exists(drive_letter + ":/.drive") {
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

func AddDrive(drive_letter string) string {
	if !helper.Exists(drive_letter + ":/.drive") {
		ksuid := helper.GenKsuid()

		drive_info := db.DriveInDB(ksuid)

		if drive_info == "" {
			id := db.InsertDriveDB(ksuid)

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
