package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var commands = map[string]int{
	"Add new path.":     1,
	"List saved drives": 2,
	"Quit":              3,
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

/*func load_db() *sql.DB {
	db, _ := sql.Open("sqlite3", "./sqlite-database.db") // Open the created SQLite File
	defer db.Close()                                     // Defer Closing the database
	return db
}*/

func help() {
	for k, v := range commands {
		fmt.Println(k + " - " + strconv.Itoa(v))
	}
}

func execute_sql(db *sql.DB, sql_str string) {
	fmt.Println("Executing SQL.")
	stmt, err := db.Prepare(sql_str) // Prepare SQL Statement
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Exec() // Execute SQL Statements
	fmt.Println("SQL query was executed")
}

func execute_sql_query(db *sql.DB, sql_str string) *sql.Rows {
	fmt.Println("Executing SQL query.")
	rows, _ := db.Query(sql_str)
	return rows
}

func main() {
	create_db()
	db, err := sql.Open("sqlite3", "./sqlite-database.db") // Open the created SQLite File

	if err != nil {
		fmt.Println(err)
	}

	defer db.Close() // Defer Closing the database

	execute_sql(db, `CREATE TABLE IF NOT EXISTS drives(
		'id' integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		'drive_letter' TEXT
	);`)

	help()

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("-----------------------------")
	fmt.Println("Type a number assigned to command")

	input := bufio.NewScanner(os.Stdin)

	for input.Scan() {

		indata := input.Text()

		fmt.Println("Input key is: " + indata)

		if indata == "1" {
			fmt.Println("Type a path for new drive")
			input_2 := bufio.NewScanner(os.Stdin)

			for input_2.Scan() {
				indata := input_2.Text()

				fmt.Println(indata)

				/*err = db.add_drive(DriveStruct{Path: filepath.Clean(indata)})
				if err != nil {
					fmt.Printf("Drive couldnt be saved because: %s \n", err)
				}*/
				break
			}

		} else if indata == "2" {
			drives := get_drives()

			if err == nil {
				for _, d := range drives {
					fmt.Println(d)
				}
				if len(drives) == 0 {
					fmt.Println("There arent any saved drives")
				}
			} else {
				fmt.Println(err)
			}

		} else if indata == "3" {
			break
		}

		fmt.Println("Type a number assigned to command")
		fmt.Println("-----------------------------")
	}

	//fmt.Println(getdrives())
}
