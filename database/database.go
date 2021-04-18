package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	helper "github.com/Fancman/BackupSoftware/helpers"
	"github.com/jamiefdhurst/journal/pkg/database/rows"
)

/*type Database interface {
	Query(sql string, args ...interface{}) (sql.Rows, error)
	Exec(sql string, args ...interface{}) (sql.Result, error)
	OpenConnection() error
}*/

type SQLite struct {
	//Database
	db *sql.DB
}

func (conn *SQLite) QueryRows(sql string, args ...interface{}) (rows.Rows, error) {
	stmt, err := conn.db.Prepare(sql)

	if err != nil {
		fmt.Println(err)
	}

	defer stmt.Close()

	rows, err := stmt.Query(args...)

	if err != nil {
		fmt.Println(err)
	}

	defer rows.Close()

	return rows, nil
}

func (conn *SQLite) Exec(sql string, args ...interface{}) (sql.Result, error) {
	tx, err := conn.db.Begin()

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	stmt, err := tx.Prepare(sql)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	result, err := stmt.Exec(args...)
	if err != nil {
		fmt.Println(err.Error())
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return result, nil
}

// Removes record from drives table
func (conn *SQLite) DelDriveDB(ksuid string) bool {
	err := conn.OpenConnection()

	if err != nil {
		return false
	}

	_, err = conn.db.Exec(`DELETE FROM drive WHERE drive_ksuid=?`, ksuid)

	if err != nil {
		return false
	}

	return true
}

func (conn *SQLite) CreateArchive(archive_name string) int64 {

	err := conn.OpenConnection()

	if err != nil {
		return 0
	}

	result, err := conn.db.Exec(`INSERT INTO archive(name) VALUES (?)`, archive_name)

	if err != nil {
		return 0
	}

	id, err := result.LastInsertId()

	return id
}

func (conn *SQLite) CreateSource(drive_ksuid string, path string) int64 {

	err := conn.OpenConnection()

	if err != nil {
		return 0
	}

	result, err := conn.db.Exec(`INSERT INTO source(drive_ksuid, path) VALUES (?, ?)`, drive_ksuid, path)

	if err != nil {
		fmt.Println(err)
		return 0
	}

	id, err := result.LastInsertId()

	return id
}

func (conn *SQLite) CreateBackup(archive_id int64, drive_ksuid string, path string) int64 {

	err := conn.OpenConnection()

	if err != nil {
		return 0
	}

	result, err := conn.db.Exec(`INSERT INTO backup(archive_id, drive_ksuid, path) VALUES (?, ?, ?)`, archive_id, drive_ksuid, path)

	if err != nil {
		return 0
	}

	id, err := result.LastInsertId()

	return id
}

func (conn *SQLite) UpdateSourceArchive(source_id int64, archive_id int64) int64 {

	err := conn.OpenConnection()

	if err != nil {
		return 0
	}

	result, err := conn.db.Exec(`UPDATE source SET archive_id = ? where id = ?`, archive_id, source_id)

	if err != nil {
		return 0
	}

	id, err := result.LastInsertId()

	return id
}

// Creates database if doesnt exist
func CreateDB() (string, error) {
	appdata_path, err := helper.GetAppDir()

	if err != nil {
		fmt.Println(err.Error())
		return appdata_path, err
	}

	var database_path = appdata_path + "/BackupSoft/sqlite-database.db"

	if !helper.Exists(database_path) {
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

func (conn *SQLite) Fixtures() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS "drive" (
			drive_ksuid VARCHAR NOT NULL,
			name VARCHAR,
			CONSTRAINT drive_PK PRIMARY KEY (drive_ksuid)
		);`,
		`CREATE TABLE IF NOT EXISTS "archive" (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR
		);`,
		`CREATE TABLE IF NOT EXISTS "backup" (
			archive_id INTEGER,
			drive_ksuid VARCHAR,
			"path" VARCHAR,
			CONSTRAINT backup_PK PRIMARY KEY (archive_id,drive_ksuid),
			CONSTRAINT backup_FK FOREIGN KEY (archive_id) REFERENCES archive(id) ON DELETE RESTRICT ON UPDATE CASCADE,
			CONSTRAINT backup_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid) ON DELETE RESTRICT ON UPDATE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS "source" (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			archive_id INTEGER,
			drive_ksuid VARCHAR,
			"path" VARCHAR,
			CONSTRAINT source_FK FOREIGN KEY (archive_id) REFERENCES archive(id) ON DELETE RESTRICT ON UPDATE CASCADE,
			CONSTRAINT source_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid) ON DELETE RESTRICT ON UPDATE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS "timestamp" (
			source_id INTEGER,
			drive_ksuid VARCHAR,
			created_at timestamp DEFAULT (strftime('%s', 'now')) NOT NULL,
			CONSTRAINT timestamp_PK PRIMARY KEY (source_id,drive_ksuid),
			CONSTRAINT timestamp_FK FOREIGN KEY (source_id) REFERENCES "source"(id),
			CONSTRAINT timestamp_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid)
		);`,
	}

	err := conn.OpenConnection()

	if err != nil {
		return err
	}

	for _, table := range tables {
		_, err := conn.db.Exec(table)

		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}

func (conn *SQLite) OpenConnection() error {
	database_path, err := CreateDB()

	if err != nil {
		fmt.Println(err)
	}

	conn.db, err = sql.Open("sqlite3", database_path) // Open the created SQLite File

	if err != nil {
		fmt.Println(err)
	}

	return conn.db.Ping()
}

// Inserts record into table drives
func (conn *SQLite) InsertDriveDB(ksuid string) int64 {

	err := conn.OpenConnection()

	if err != nil {
		return 0
	}

	result, err := conn.db.Exec(
		`INSERT INTO drive(drive_ksuid) VALUES (?)`, ksuid)

	if err != nil {
		return 0
	}

	id, err := result.LastInsertId()

	return id
}

// Returns drive by ksuid from db
func (conn *SQLite) DriveInDB(drive_ksuid string) string {
	err := conn.OpenConnection()

	if err != nil {
		return ""
	}

	stmt := `SELECT drive_ksuid FROM drive
	WHERE drive_ksuid=?`
	var r_ksuid string

	row := conn.db.QueryRow(stmt, drive_ksuid)
	switch err := row.Scan(&r_ksuid); err {
	case nil:
		return r_ksuid
	default:
		return ""
	}
}

// Inserts record into drives table
/*func insert_backups_db(source string, dest_drive_ksuid string, dest_drive_letter string, path string) error {
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
*/
