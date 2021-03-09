package cmd

import (
	"database/sql"
	"fmt"
	"os"
)

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

func open_conn() (db *sql.DB) {
	create_db()

	db, err = sql.Open("sqlite3", "./sqlite-database.db") // Open the created SQLite File

	if err != nil {
		fmt.Println(err)
	}

	//defer db.Close() // Defer Closing the database

	return db
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
