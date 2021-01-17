package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/segmentio/ksuid"

	_ "github.com/mattn/go-sqlite3"
)

// Prikazy pri interativnom mode
var commands = map[string]int{
	"Pridat novy drive do db": 1,
	"Vypisat dostupne drivy":  2,
	"Vytvorit zalohu":         3,
	"Quit":                    4,
}

var db *sql.DB
var err error

type Backup struct {
	id                int
	source            string
	destination_ksuid string
	cron              string
}

// Vrati pole zaznamov z tabulky backups
func find_backups(db *sql.DB) []Backup {
	var id int
	var source string
	var destination_ksuid string
	var cron string

	rows, err := db.Query(`SELECT id, source, destination_ksuid, cron FROM backups;`)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var backups []Backup

	for rows.Next() {
		err := rows.Scan(&id, &source, &destination_ksuid, &cron)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(id, source, destination_ksuid, cron)
		backups = append(backups, Backup{id: id, source: source, destination_ksuid: destination_ksuid, cron: cron})
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	return backups
}

// Vrati drive podla pismena
func drive_db_exists_letter(drive_letter string) (bool, []string) {
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
}

// Vrati drive podla ksuid
func drive_db_exists_ksuid(drive_ksuid string) (bool, []string) {
	stmt := `SELECT drive_letter, drive_ksuid FROM drives 
	WHERE drive_ksuid=?`
	var r_ksuid string
	var r_letter string
	row := db.QueryRow(stmt, drive_ksuid)
	switch err := row.Scan(&r_letter, &r_ksuid); err {
	case sql.ErrNoRows:
		return false, []string{}
	case nil:
		return true, []string{r_letter, r_ksuid}
	default:
		panic(err)
	}
}

// Vrati zoznam dostypnych diskov
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

// Existuje súbor?
func file_exists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Vytvori databazu ak neexistuje
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

// Vypise vsetky prikazy pri interaktivnom rezime
func help() {
	for k, v := range commands {
		fmt.Println(k + " - " + strconv.Itoa(v))
	}
}

// Spusti SQL prikaz
func execute_sql(sql_str string) {
	fmt.Println("Executing SQL.")
	stmt, err := db.Prepare(sql_str) // Prepare SQL Statement
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Exec() // Execute SQL Statements
	fmt.Println("SQL query was executed")
}

// Spusti SQL query a vrati jej riadky
func execute_sql_query(sql_str string) *sql.Rows {
	fmt.Println("Executing SQL query.")
	rows, _ := db.Query(sql_str)
	return rows
}

// Vlozi zaznam do tabulky drives
func insert_drive_db(drive_letter string, ksuid string) {
	sql_str := `INSERT INTO drives(drive_letter, drive_ksuid) 
	VALUES (?, ?)`
	stmt, err := db.Prepare(sql_str) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = stmt.Exec(drive_letter, ksuid)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Vymaze zaznam z tabulky drives
func delete_drive_db(ksuid string) {
	fmt.Println("Mazanie drivu z tabulky")
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

// Vlozi zaznam do tabulky backups
func insert_backups_db(source string, destination string, cron string) {
	exists, info := drive_db_exists_letter(string(destination))

	if exists {
		sql_str := `INSERT INTO backups(source, destination_ksuid, cron) 
		VALUES (?, ?, ?)`
		stmt, err := db.Prepare(sql_str) // Prepare statement.
		// This is good to avoid SQL injections
		if err != nil {
			fmt.Println(err.Error())
		}
		_, err = stmt.Exec(source, info[1], cron)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println("Disk nie je v databaze")
	}
}

// Cita riadky z txt suboru
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

// Zapise na drive .drive subor s ksuid
func write_disk_identification(drive_letter string, ksuid string) {
	currentTime := time.Now()

	data := []string{
		drive_letter,
		ksuid,
		currentTime.String(),
	}

	file, err := os.OpenFile(drive_letter+":/.drive", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	fmt.Println("Vytvaranie .drive suboru na " + drive_letter + ":/.drive")
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

// Testujem kompresiu
func start_backup(source string, destination_ksuid string) {
	exists, info := drive_db_exists_ksuid(string(destination_ksuid))
	if exists {
		args := []string{"a", "-t7z", info[0] + ":/backup/test.7zip", source}
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

// Vypise zoznam driveov a ich stav
func list_drives() {
	drives := get_drives()

	if len(drives) > 0 {
		for _, d := range drives {
			exists, info := drive_db_exists_letter(string(d))

			if file_exists(d + ":/.drive") {
				// ak ma .drive subor a nie je zapisane v db
				if !exists {
					lines, err := read_file_lines(d + ":/.drive")
					if err != nil {
						fmt.Printf("Error pri citani suboru: %s", err)
					}

					insert_drive_db(lines[0], lines[1])
				}
				// ak existuju obidve osetrit ak sa nezhoduju udaje
			} else {
				// ak nema .drive subor a je zapisane v db
				if exists {
					write_disk_identification(info[0], info[1])
				}
			}
			fmt.Print(d)

			if exists {
				fmt.Print(" - ulozene pod ksid" + info[1])
			}
			fmt.Print("\n")
		}
	} else {
		fmt.Println("Nie su ziadne disky na PC")
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

	backups := find_backups(db)

	if len(backups) > 0 {
		for _, b := range backups {
			start_backup(b.source, b.destination_ksuid)
		}
	}

	fmt.Println("STOP")
	os.Exit(3)

	//execute_sql(db, `DROP TABLE IF EXISTS drives;`)

	execute_sql(`CREATE TABLE IF NOT EXISTS drives(
		'id' integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		'drive_letter' TEXT,
		'drive_ksuid' TEXT
	);`)

	execute_sql(`CREATE TABLE IF NOT EXISTS backups(
		'id' integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		'source' TEXT,
		'destination_ksuid' TEXT,
		'cron' TEXT
	);`)

	help()

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("-----------------------------")
	fmt.Println("Zadaj cislo priradene k prikazu")

	input := bufio.NewScanner(os.Stdin)

	for input.Scan() {

		indata := input.Text()

		fmt.Println("Zadane cislo  je: " + indata)

		if indata == "1" {
			list_drives()
			fmt.Println("Zadaj pismeno priradene k disku")
			input_2 := bufio.NewScanner(os.Stdin)

			for input_2.Scan() {
				indata := input_2.Text()

				drives := get_drives()
				drive_exists := false

				if len(drives) > 0 {
					for _, d := range drives {
						if d == indata {
							drive_exists = true
						}
					}
				}

				//fmt.Println(indata)

				if drive_exists {
					exists, info := drive_db_exists_letter(string(indata))
					ksuid := gen_ksuid()
					if !exists {
						insert_drive_db(indata, ksuid)
						write_disk_identification(indata, ksuid)
					} else {
						fmt.Println("Drive s rovnakym pismenom je uz v db pod drive_ksuid" + info[1])
					}
				} else {
					fmt.Println("Drive so zadanym pismenom neexistuje")
				}

				/*err = db.add_drive(DriveStruct{Path: filepath.Clean(indata)})
				if err != nil {
					fmt.Printf("Drive couldnt be saved because: %s \n", err)
				}*/
				break
			}

		} else if indata == "2" {
			list_drives()
		} else if indata == "3" {
			fmt.Println("Zadaj priecinok na zalohu")
			input_2 := bufio.NewScanner(os.Stdin)

			for input_2.Scan() {
				source := input_2.Text()

				fmt.Println("Zadaj pismeno cieloveho disku na zalohu")
				input_3 := bufio.NewScanner(os.Stdin)

				for input_3.Scan() {
					destination := input_3.Text()

					fmt.Println("Zadaj cron zalohy")
					input_4 := bufio.NewScanner(os.Stdin)

					for input_4.Scan() {
						cron := input_4.Text()

						fmt.Println(filepath.Clean(source) + " " + destination + " " + cron)

						insert_backups_db(filepath.Clean(source), destination, cron)

						/*err = db.add_drive(DriveStruct{Path: filepath.Clean(indata)})
						if err != nil {
							fmt.Printf("Drive couldnt be saved because: %s \n", err)
						}*/
						break
					}
					break
				}
				break
			}
			//break
		} else if indata == "4" {
			break
		}

		fmt.Println("Zadaj cislo priradene k prikazu")
		fmt.Println("-----------------------------")
	}

	//fmt.Println(getdrives())
}
