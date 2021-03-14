package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

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
func write_disk_identification(drive_letter string, ksuid string) {
	currentTime := time.Now()

	data := []string{
		ksuid,
		currentTime.String(),
	}

	file, err := os.OpenFile(drive_letter+":/.drive", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	//fmt.Println("Creating .drive file on th path: " + drive_letter + ":/.drive")
	if err != nil {
		fmt.Printf("Nepodarilo sa vytvorit subor %s\n", err)
		delete_drive_db(ksuid)
	}

	datawriter := bufio.NewWriter(file)

	for _, data := range data {
		_, _ = datawriter.WriteString(data + "\n")
	}

	datawriter.Flush()
	file.Close()
}

func start_restore(id int, source string, destinations []Destination) error {
	if len(destinations) == 1 {
		exists, db_drive_ksuid := drive_db_exists_ksuid(destinations[0].ksuid)

		if exists {
			drive_letter := path2drive(db_drive_ksuid)

			err := os.MkdirAll(filepath.Dir(source), os.ModePerm)

			//fmt.Println(filepath.Dir(source))

			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			dest_path := drive_letter + ":/" + destinations[0].path

			fmt.Printf("Cesta zalohy: %s", dest_path)

			_, err = os.Stat(dest_path + "/" + strconv.Itoa(id) + ".7z")

			if os.IsNotExist(err) {
				return err
			}

			_, err = os.Stat("7-ZipPortable/App/7-Zip64/7z.exe")

			if os.IsNotExist(err) {
				return err
			}

			args := []string{"x", dest_path + "/" + strconv.Itoa(id) + ".7z", "-y", "-o" + source}

			cmd := exec.Command("7-ZipPortable/App/7-Zip64/7z.exe", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			/*err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
				fmt.Println("Destination: " + dest_path + "/" + info.Name() + ".7z")
				fmt.Println("Source: " + source)

				args := []string{"a", "-t7z", dest_path + "/" + info.Name() + ".7z", source}

				cmd := exec.Command("7-ZipPortable/App/7-Zip64/7z.exe", args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				if err != nil {
					fmt.Println(err.Error())
				}

				return nil
			})*/

			if err != nil {
				fmt.Println(err.Error())
				return err
			}

		}
	}

	return nil
}

// Testing compression
func start_backup(id int, source string, destinations []Destination) error {
	if len(destinations) == 1 {
		exists, db_drive_ksuid := drive_db_exists_ksuid(destinations[0].ksuid)

		if exists {
			drive_letter := path2drive(db_drive_ksuid)
			//dt := time.Now()

			dest_path := drive_letter + ":/" + destinations[0].path

			fmt.Printf("Cesta zalohy: %s", dest_path)

			err := os.MkdirAll(dest_path, os.ModePerm)

			if err != nil {
				fmt.Println(err.Error())
			}

			// strconv.Itoa(id) nazov archivu

			info, err := os.Stat(source)
			if os.IsNotExist(err) {
				fmt.Println("File does not exist.")
			}
			if info.IsDir() {
				fmt.Println("temp is a directory")
			} else {
				fmt.Println("temp is a file")
			}

			info, err = os.Stat("7-ZipPortable/App/7-Zip64/7z.exe")

			if os.IsNotExist(err) {
				return err
			}

			args := []string{"a", "-t7z", dest_path + "/" + strconv.Itoa(id) + ".7z", source + "/*"}

			cmd := exec.Command("7-ZipPortable/App/7-Zip64/7z.exe", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			/*err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
				fmt.Println("Destination: " + dest_path + "/" + info.Name() + ".7z")
				fmt.Println("Source: " + source)

				args := []string{"a", "-t7z", dest_path + "/" + info.Name() + ".7z", source}

				cmd := exec.Command("7-ZipPortable/App/7-Zip64/7z.exe", args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				if err != nil {
					fmt.Println(err.Error())
				}

				return nil
			})*/

			if err != nil {
				fmt.Println(err.Error())
				return err
			}

		}
	}

	return nil
}

// Lists drives and their statuses
func list_drives() {
	drives := get_drives()

	if len(drives) > 0 {
		for _, drive_letter := range drives {

			fmt.Print(drive_letter)

			if file_exists(drive_letter + ":/.drive") {
				// ak ma .drive subor a nie je zapisane v db
				lines, err := read_file_lines(drive_letter + ":/.drive")

				if err != nil {
					fmt.Printf("Error pri citani suboru: %s", err)
				}

				exists, info := drive_db_exists_ksuid(string(lines[0]))
				// I have to use this because I cant have unused variables
				//_ = info

				// If .drive exists but isnt in db
				fmt.Print(" - " + string(lines[0]))
				if !exists && info == "" {
					insert_drive_db(string(lines[0]))
					fmt.Print(" - Drive has .drive file but werent in drives table.")
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

// Get path to drive by ksuid
func path2drive(ksuid string) string {
	drives := get_drives()

	if len(drives) > 0 {
		for _, drive_letter := range drives {

			//fmt.Print(drive_letter)

			if file_exists(drive_letter + ":/.drive") {
				// ak ma .drive subor a nie je zapisane v db
				lines, err := read_file_lines(drive_letter + ":/.drive")

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

func add_drive(drive_letter string) {
	if !file_exists(drive_letter + ":/.drive") {
		ksuid := gen_ksuid()

		exists, info := drive_db_exists_ksuid(ksuid)

		if !exists && info == "" {
			insert_drive_db(ksuid)
			write_disk_identification(drive_letter, ksuid)
		}
	}

}
