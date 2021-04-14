package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
func create_db() (string, error) {
	appdata_path, err := get_appdata_dir()

	if err != nil {
		fmt.Println(err.Error())
		return appdata_path, err
	}

	var database_path = appdata_path + "/BackupSoft/sqlite-database.db"

	fmt.Println(database_path)

	if !file_exists(database_path) {
		err := os.MkdirAll(filepath.Dir(database_path), os.ModePerm)

		if err != nil {
			fmt.Println(err.Error())
			return database_path, err
		}

		file, err := os.Create(database_path)
		if err != nil {
			fmt.Println(err.Error())
			return database_path, err
		}

		file.Close()
		fmt.Println("sqlite-database.db created")

	}

	return database_path, nil
}

func open_conn() (db *sql.DB) {
	database_path, err := create_db()

	if err != nil {
		fmt.Println(err)
	}

	db, err = sql.Open("sqlite3", database_path) // Open the created SQLite File

	if err != nil {
		fmt.Println(err)
	}

	//defer db.Close() // Defer Closing the database
	execute_sql(db, `CREATE TABLE IF NOT EXISTS drive (
		drive_ksuid INTEGER NOT NULL,
		name VARCHAR,
		CONSTRAINT drive_PK PRIMARY KEY (drive_ksuid)
	);
	CREATE INDEX drive_drive_ksuid_IDX ON drive (drive_ksuid);`)

	execute_sql(db, `CREATE TABLE IF NOT EXISTS archive (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR
	);`)

	execute_sql(db, `CREATE TABLE IF NOT EXISTS backup (
		archive_id INTEGER,
		drive_ksuid INTEGER,
		"path" VARCHAR,
		CONSTRAINT backup_PK PRIMARY KEY (archive_id,drive_ksuid),
		CONSTRAINT backup_FK FOREIGN KEY (archive_id) REFERENCES archive(id) ON DELETE RESTRICT ON UPDATE CASCADE,
		CONSTRAINT backup_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid) ON DELETE RESTRICT ON UPDATE CASCADE
	);`)

	execute_sql(db, `CREATE TABLE IF NOT EXISTS "source" (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		archive_id INTEGER,
		drive_ksuid INTEGER,
		"path" VARCHAR,
		CONSTRAINT source_FK FOREIGN KEY (id) REFERENCES archive(id) ON DELETE RESTRICT ON UPDATE CASCADE,
		CONSTRAINT source_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid) ON DELETE RESTRICT ON UPDATE CASCADE
	);`)

	execute_sql(db, `CREATE TABLE IF NOT EXISTS "timestamp" (
		source_id INTEGER,
		drive_ksuid INTEGER,
		created_at timestamp DEFAULT (strftime('%s', 'now')) NOT NULL,
		CONSTRAINT timestamp_PK PRIMARY KEY (source_id,drive_ksuid),
		CONSTRAINT timestamp_FK FOREIGN KEY (source_id) REFERENCES "source"(id),
		CONSTRAINT timestamp_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid)
	);`)

	return db
}

// Executes SQL commans
func execute_sql(db *sql.DB, sql_str string) {
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
func insert_backups_db(source string, dest_drive_ksuid string, dest_drive_letter string, path string) error {
	db_drive_ksuid := ""
	exists := false

	if len(dest_drive_letter) != 0 {
		err := drive_exists(string(dest_drive_letter))
		if err == nil {
			db_drive_ksuid, err = get_ksuid_from_drive(dest_drive_letter)
			exists = true
		}
	} else if len(dest_drive_ksuid) != 0 {
		exists, db_drive_ksuid = drive_db_exists_ksuid(dest_drive_ksuid)
	} else {
		return errors.New("No input argument for backup destination")
	}

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
