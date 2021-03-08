package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/segmentio/ksuid"

	_ "github.com/mattn/go-sqlite3"
)

// Commands in interactive mod
var commands = map[string]int{
	"Pridat novy drive do db": 1,
	"Vypisat dostupne drivy":  2,
	"Vytvorit zalohu":         3,
	"Quit":                    4,
}

var db *sql.DB
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

// Returns drive by ksuid from db
func drive_db_exists_ksuid(drive_ksuid string) (bool, string) {
	stmt := `SELECT drive_ksuid FROM drives 
	WHERE drive_ksuid=?`
	var r_ksuid string

	row := db.QueryRow(stmt, drive_ksuid)
	switch err := row.Scan(&r_ksuid); err {
	case sql.ErrNoRows:
		return false, ""
	case nil:
		return true, r_ksuid
	default:
		panic(err)
	}
}

// Returns list of available drives
func get_drives() (r []string) {
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		f, err := os.Open(string(drive) + ":\\")
		if err == nil {
			r = append(r, string(drive))
			f.Close()
		}
	}
	return r
}

// File exists?
func file_exists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Creates database if doesnt exist
func create_db() {
	//fmt.Println("Create sqlite db")
	if !file_exists("sqlite-database.db") {
		file, err := os.Create("sqlite-database.db")
		if err != nil {
			fmt.Println(err.Error())
		}
		file.Close()
		fmt.Println("sqlite-database.db created")
	} else {
		//fmt.Println("sqlite-database.db already exists")
	}
}

// List all commands in interactive mod
func help() {
	for k, v := range commands {
		fmt.Println(k + " - " + strconv.Itoa(v))
	}
}

// Executes SQL commans
func execute_sql(sql_str string) {
	//fmt.Println("Executing SQL.")
	stmt, err := db.Prepare(sql_str) // Prepare SQL Statement
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Exec() // Execute SQL Statements
	//fmt.Println("SQL query was executed")
}

// Executes SQL commans and returns its rows
func execute_sql_query(sql_str string) *sql.Rows {
	//fmt.Println("Executing SQL query.")
	rows, _ := db.Query(sql_str)
	return rows
}

// Inserts record into table drives
func insert_drive_db(ksuid string) {
	sql_str := `INSERT INTO drives(drive_ksuid) 
	VALUES (?)`
	stmt, err := db.Prepare(sql_str)

	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = stmt.Exec(ksuid)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Removes record from drives table
func delete_drive_db(ksuid string) {
	//fmt.Println("Delete from drives table by ksuid")
	sql_str := `DELETE FROM drives WHERE drive_ksuid=?`
	stmt, err := db.Prepare(sql_str) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = stmt.Exec(ksuid)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Inserts record into drives table
func insert_backups_db(source string, dest_drive_ksuid string, path string) error {
	exists, db_drive_ksuid := drive_db_exists_ksuid(dest_drive_ksuid)

	if exists {
		sql_str := `INSERT INTO backups(source) 
		VALUES (?)`
		stmt, err := db.Prepare(sql_str) // Prepare statement.
		// This is good to avoid SQL injections
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		res, err := stmt.Exec(source)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		id, err := res.LastInsertId()

		if err != nil {
			return err
		}

		sql_str = `INSERT INTO destinations(backup_id, drive_ksuid, path)
		VALUES (?, ?, ?)`
		stmt, err = db.Prepare(sql_str)

		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		_, err = stmt.Exec(id, db_drive_ksuid, path)

		if err != nil {
			fmt.Println(err.Error())
			return err
		}

	} else {
		fmt.Println("Disk nie je v databaze")
	}

	return nil
}

// Reads lines from text file
func read_file_lines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

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

// https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html
func gen_ksuid() string {
	return ksuid.New().String()
	//fmt.Printf("github.com/segmentio/ksuid:     %s\n", id.String())
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

			args := []string{"a", "-t7z", dest_path + "/" + strconv.Itoa(id) + ".7z", source}

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

// main funkcia
func main() {
	create_db()
	db, err = sql.Open("sqlite3", "./sqlite-database.db") // Open the created SQLite File

	if err != nil {
		fmt.Println(err)
	}

	defer db.Close() // Defer Closing the database

	execute_sql(`CREATE TABLE IF NOT EXISTS drives(
		'id' integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		'drive_ksuid' TEXT
	);`)

	execute_sql(`CREATE TABLE IF NOT EXISTS backups(
		'id' integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		'source' TEXT
	);`)

	execute_sql(`CREATE TABLE IF NOT EXISTS destinations (
		backup_id INTEGER NOT NULL,
		drive_ksuid INTEGER NOT NULL,
		'path' TEXT
	);`)

	//insert_backups_db(path.Clean("C:/Users/tomas/Pictures/Backgrounds"), "1mC60uVtv07vvPY4ylFkaXlc4b9", path.Clean("backup"))

	//list_drives()

	//add_drive("F")

	//insert_backups_db(source string, dest_drive_ksuid string, path string)

	backups := find_backups(db, 3)

	fmt.Println("floor")

	if len(backups) > 0 {
		for _, b := range backups {
			//fmt.Printf("Backup is: %b %s %v", b.id, b.source, b.destinations)
			start_backup(b.id, b.source, b.destinations)
		}
	}

	//fmt.Println("STOP")
	//os.Exit(3)

	//execute_sql(db, `DROP TABLE IF EXISTS drives;`)

	//fmt.Println(getdrives())
}
