package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
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

type Backup struct {
	id           int
	source       string
	destinations []string
}

// Returns records from table 'backups'
func find_backups(db *sql.DB) []Backup {
	var id int
	var source string
	var destination_ksuid string
	//var cron string

	rows, err := db.Query(`SELECT b.id, b.source, dr.drive_ksuid FROM backups b JOIN destinations de ON de.backup_id = b.id JOIN drives dr ON dr.id=de.drive_ksuid;`)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var backups []Backup

	for rows.Next() {
		var exists_in_backups = false

		err := rows.Scan(&id, &source, &destination_ksuid)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(id, source, destination_ksuid)

		for k, v := range backups {
			if v.id == id {
				backups[k].destinations = append(backups[k].destinations, destination_ksuid)
				exists_in_backups = true
				break
			}
		}
		if !exists_in_backups {
			backups = append(backups, Backup{id: id, source: source, destinations: []string{destination_ksuid}})
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
	fmt.Println("Create sqlite db")
	if !file_exists("sqlite-database.db") {
		file, err := os.Create("sqlite-database.db")
		if err != nil {
			fmt.Println(err.Error())
		}
		file.Close()
		fmt.Println("sqlite-database.db created")
	} else {
		fmt.Println("sqlite-database.db already exists")
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
	fmt.Println("Executing SQL.")
	stmt, err := db.Prepare(sql_str) // Prepare SQL Statement
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Exec() // Execute SQL Statements
	fmt.Println("SQL query was executed")
}

// Executes SQL commans and returns its rows
func execute_sql_query(sql_str string) *sql.Rows {
	fmt.Println("Executing SQL query.")
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
	fmt.Println("Delete from drives table by ksuid")
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
func insert_backups_db(source string, dest_drive_ksuid string, cron string) {
	exists, db_drive_ksuid := drive_db_exists_ksuid(dest_drive_ksuid)

	if exists {
		sql_str := `INSERT INTO backups(source, destination_ksuid, cron) 
		VALUES (?, ?, ?)`
		stmt, err := db.Prepare(sql_str) // Prepare statement.
		// This is good to avoid SQL injections
		if err != nil {
			fmt.Println(err.Error())
		}
		_, err = stmt.Exec(source, db_drive_ksuid, cron)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println("Disk nie je v databaze")
	}
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
	fmt.Println("Creating .drive file on th path: " + drive_letter + ":/.drive")
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
func start_backup(source string, destination_ksuid string) {
	exists, db_drive_ksuid := drive_db_exists_ksuid(string(destination_ksuid))
	if exists {
		args := []string{"a", "-t7z", db_drive_ksuid + ":/backup/test.7zip", source}
		//" a -t7z '"+info[0]+":/backup/test.7zip' '"+source+"'"
		//args..
		//fmt.Println("a -t7z " + info[0] + ":\backup\test.7zip " + source)
		cmd := exec.Command("7-ZipPortable/App/7-Zip64/7z.exe", args...)
		//cmd := exec.Command("ls", "-lah")
		//cmd := exec.Command("C:/Users/fancy/go/src/bakalarska praca/7z/7z-portable.exe", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		/*if err := cmd.Start(); err != nil {
			fmt.Println(err)
		}*/
	}
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

			fmt.Print(drive_letter)

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
		'source' TEXT,
		'cron' TEXT
	);`)

	execute_sql(`CREATE TABLE IF NOT EXISTS destinations (
		backup_id INTEGER NOT NULL,
		drive_ksuid INTEGER NOT NULL
	);`)

	list_drives()

	add_drive("D")

	backups := find_backups(db)

	if len(backups) > 0 {
		for _, b := range backups {
			fmt.Printf("Backup is: %b %s %v", b.id, b.source, b.destinations)
			//start_backup(b.source, b.destination_ksuid)
		}
	}

	//fmt.Println("STOP")
	//os.Exit(3)

	//execute_sql(db, `DROP TABLE IF EXISTS drives;`)

	//fmt.Println(getdrives())
}
