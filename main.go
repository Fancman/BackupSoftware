package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/segmentio/ksuid"

	_ "github.com/mattn/go-sqlite3"
)

var commands = map[string]int{
	"Pridat novy drive do db": 1,
	"Vypisat dostupne drivy":  2,
	"Vytvorit zalohu":         3,
	"Quit":                    4,
}

var db *sql.DB
var err error

func drive_db_exists(drive_letter string) (bool, []string) {
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

func file_exists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

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

func help() {
	for k, v := range commands {
		fmt.Println(k + " - " + strconv.Itoa(v))
	}
}

func execute_sql(sql_str string) {
	fmt.Println("Executing SQL.")
	stmt, err := db.Prepare(sql_str) // Prepare SQL Statement
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Exec() // Execute SQL Statements
	fmt.Println("SQL query was executed")
}

func execute_sql_query(sql_str string) *sql.Rows {
	fmt.Println("Executing SQL query.")
	rows, _ := db.Query(sql_str)
	return rows
}

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

func insert_backups_db(source string, destination string, cron string) {
	exists, info := drive_db_exists(string(destination))

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
		fmt.Println("Nepodarilo sa vytvorit subor")
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

func list_drives() {
	drives := get_drives()

	if len(drives) > 0 {
		for _, d := range drives {
			exists, info := drive_db_exists(string(d))

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

func main() {
	create_db()
	db, err = sql.Open("sqlite3", "./sqlite-database.db") // Open the created SQLite File

	if err != nil {
		fmt.Println(err)
	}

	defer db.Close() // Defer Closing the database

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

				//fmt.Println(indata)

				exists, info := drive_db_exists(string(indata))
				ksuid := gen_ksuid()
				if !exists {
					insert_drive_db(indata, ksuid)
					write_disk_identification(indata, ksuid)
				} else {
					fmt.Println("Drive s rovnakym pismenom je uz v db pod drive_ksuid" + info[1])
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
			break
		} else if indata == "4" {
			break
		}

		fmt.Println("Zadaj cislo priradene k prikazu")
		fmt.Println("-----------------------------")
	}

	//fmt.Println(getdrives())
}
