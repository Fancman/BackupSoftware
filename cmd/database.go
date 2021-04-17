package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Removes record from drives table
func DelDriveDB(ksuid string) bool {
	tx, err := db.Begin()

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	sql_str := `DELETE FROM drive WHERE drive_ksuid=?`
	stmt, err := tx.Prepare(sql_str)

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	_, err = stmt.Exec(ksuid)
	if err != nil {
		fmt.Println(err.Error())
		tx.Rollback()
		return false
	}

	tx.Commit()
	return true
}

func CreateArchive(archive_name string) int64 {
	sql_str := `INSERT INTO archive(name) VALUES (?)`

	tx, err := db.Begin()
	if err != nil {
		return 0
	}

	stmt, err := tx.Prepare(sql_str)

	if err != nil {
		fmt.Println(err.Error())
		return 0
	}

	defer stmt.Close()

	res, err := stmt.Exec(archive_name)

	if err != nil {
		fmt.Println(err.Error())
		tx.Rollback()
		return 0
	}

	tx.Commit()

	id, err := res.LastInsertId()

	return id
}

// Creates database if doesnt exist
func CreateDB() (string, error) {
	appdata_path, err := get_appdata_dir()

	if err != nil {
		fmt.Println(err.Error())
		return appdata_path, err
	}

	var database_path = appdata_path + "/BackupSoft/sqlite-database.db"

	fmt.Println(database_path)

	if !FileExists(database_path) {
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

func OpenConnection() (db *sql.DB) {
	database_path, err := CreateDB()

	if err != nil {
		fmt.Println(err)
	}

	db, err = sql.Open("sqlite3", database_path) // Open the created SQLite File

	if err != nil {
		fmt.Println(err)
	}

	//defer db.Close() // Defer Closing the database
	ExecuteSQL(db, `CREATE TABLE IF NOT EXISTS "drive" (
		drive_ksuid VARCHAR NOT NULL,
		name VARCHAR,
		CONSTRAINT drive_PK PRIMARY KEY (drive_ksuid)
	);
	CREATE INDEX drive_drive_ksuid_IDX ON drive (drive_ksuid);`)

	ExecuteSQL(db, `CREATE TABLE IF NOT EXISTS "archive" (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR
	);`)

	ExecuteSQL(db, `CREATE TABLE IF NOT EXISTS "backup" (
		archive_id INTEGER,
		drive_ksuid VARCHAR,
		"path" VARCHAR,
		CONSTRAINT backup_PK PRIMARY KEY (archive_id,drive_ksuid),
		CONSTRAINT backup_FK FOREIGN KEY (archive_id) REFERENCES archive(id) ON DELETE RESTRICT ON UPDATE CASCADE,
		CONSTRAINT backup_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid) ON DELETE RESTRICT ON UPDATE CASCADE
	);`)

	ExecuteSQL(db, `CREATE TABLE IF NOT EXISTS "source" (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		archive_id INTEGER,
		drive_ksuid VARCHAR,
		"path" VARCHAR,
		CONSTRAINT source_FK FOREIGN KEY (archive_id) REFERENCES archive(id) ON DELETE RESTRICT ON UPDATE CASCADE,
		CONSTRAINT source_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid) ON DELETE RESTRICT ON UPDATE CASCADE
	);`)

	ExecuteSQL(db, `CREATE TABLE IF NOT EXISTS "timestamp" (
		source_id INTEGER,
		drive_ksuid VARCHAR,
		created_at timestamp DEFAULT (strftime('%s', 'now')) NOT NULL,
		CONSTRAINT timestamp_PK PRIMARY KEY (source_id,drive_ksuid),
		CONSTRAINT timestamp_FK FOREIGN KEY (source_id) REFERENCES "source"(id),
		CONSTRAINT timestamp_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid)
	);`)

	return db
}

// Executes SQL commans
func ExecuteSQL(db *sql.DB, sql_str string) {
	//fmt.Println("Executing SQL.")
	stmt, err := db.Prepare(sql_str) // Prepare SQL Statement
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Exec() // Execute SQL Statements
	//fmt.Println("SQL query was executed")
}

// Inserts record into table drives
func InsertDriveDB(ksuid string) bool {
	sql_str := `INSERT INTO drive(drive_ksuid) VALUES (?)`

	tx, err := db.Begin()
	if err != nil {
		return false
	}

	stmt, err := tx.Prepare(sql_str)

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	defer stmt.Close()

	_, err = stmt.Exec(ksuid)
	if err != nil {
		fmt.Println(err.Error())
		tx.Rollback()
		return false
	}

	tx.Commit()
	return true
}

// Returns drive by ksuid from db
func isDriveInDB(drive_ksuid string) (bool, string) {
	stmt := `SELECT drive_ksuid FROM drive
	WHERE drive_ksuid=?`
	var r_ksuid string

	row := db.QueryRow(stmt, drive_ksuid)
	switch err := row.Scan(&r_ksuid); err {
	case sql.ErrNoRows:
		return false, ""
	case nil:
		return true, r_ksuid
	default:
		return false, ""
	}
}

// Inserts record into drives table
func insert_backups_db(source string, dest_drive_ksuid string, dest_drive_letter string, path string) error {
	db_drive_ksuid := ""
	exists := false

	if len(dest_drive_letter) != 0 {
		err := DriveExists(string(dest_drive_letter))
		if err == nil {
			db_drive_ksuid, err = get_ksuid_from_drive(dest_drive_letter)
			exists = true
		}
	} else if len(dest_drive_ksuid) != 0 {
		exists, db_drive_ksuid = isDriveInDB(dest_drive_ksuid)
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
