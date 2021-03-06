package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	//"strings"

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
		//fmt.Println(err)
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(args...)

	if err != nil {
		//fmt.Println(err)
		return rows, err
	}

	//defer rows.Close()

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
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return false
	}

	_, err = conn.db.Exec(`DELETE FROM drive WHERE drive_ksuid=?`, ksuid)

	return err == nil
}

// Removes record from drives table
func (conn *SQLite) DelArchiveDB(archive_id int64) bool {
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return false
	}

	_, err = conn.db.Exec(`DELETE FROM archive WHERE id=?`, archive_id)

	return err == nil
}

func (conn *SQLite) RemoveDestinationByPath(archive_name string, ksuid string) bool {
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return false
	}

	res, err := conn.db.Exec(`DELETE FROM backup WHERE drive_ksuid=? AND archive_id IN (SELECT id FROM archive WHERE name=?)`, ksuid, archive_name)

	if err != nil {
		return false
	}

	affected_rows, err := res.RowsAffected()

	if err != nil {
		return false
	}

	if affected_rows == 0 {
		fmt.Println("0 rows in database were deleted.")
		return false
	}

	_, err = conn.db.Exec(`DELETE FROM archive WHERE name=? AND NOT EXISTS (SELECT * FROM backup WHERE archive_id IN (SELECT id FROM archive WHERE name=?))`, archive_name, archive_name)

	return err == nil
}

// Removes record from source table
func (conn *SQLite) RemoveSource(source_id int64) error {
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return err
	}

	var cnt int = 0

	stmt := `SELECT count(*) as cnt FROM source WHERE id = ?`

	row := conn.db.QueryRow(stmt, source_id)

	err = row.Scan(&cnt)

	if err != nil {
		return err
	}

	if cnt != 0 {
		_, err = conn.db.Exec(`DELETE FROM source WHERE id=?`, source_id)

		if err != nil {
			return err
		}

		fmt.Println("Source was removed succesfuly.")

		return nil
	}

	return errors.New("Source record with given name could not be found.")
}

// Removes record from source table
func (conn *SQLite) GetSourceArchive(source_id int64) int64 {
	var archive_id int64 = 0
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return archive_id
	}

	stmt := `SELECT archive_id FROM source WHERE id = ?`
	row := conn.db.QueryRow(stmt, source_id)
	err = row.Scan(&archive_id)

	if err != nil {
		fmt.Println(err)
	}

	return archive_id
}

// Removes record from drives table
func (conn *SQLite) RemoveDestination(archive_id int64, drive_ksuid string) bool {
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return false
	}

	_, err = conn.db.Exec(`DELETE FROM backup WHERE archive_id = ? AND drive_ksuid = ?`, archive_id, drive_ksuid)

	return err == nil
}

// Removes record from drives table
func (conn *SQLite) RemoveDestinationByDrive(drive_ksuid string) bool {
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return false
	}

	_, err = conn.db.Exec(`DELETE FROM backup WHERE drive_ksuid = ?`, drive_ksuid)

	return err == nil
}

// Removes record from drives table
func (conn *SQLite) RemoveDestinationByArchive(archive_id int64) bool {
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return false
	}

	_, err = conn.db.Exec(`DELETE FROM backup WHERE archive_id = ?`, archive_id)

	return err == nil
}

// Removes record from source table
func (conn *SQLite) RemoveSources(archive_id int64) bool {
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return false
	}

	_, err = conn.db.Exec(`DELETE FROM source WHERE archive_id = ?`, archive_id)

	return err == nil
}

// Removes record from drives table
func (conn *SQLite) RemoveDestinations(archive_id int64) bool {
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return false
	}

	_, err = conn.db.Exec(`DELETE FROM backup WHERE archive_id = ?`, archive_id)

	return err == nil
}

func (conn *SQLite) ArchiveUsed(archive_id int64) (int, int) {
	err := conn.OpenConnection(helper.GetDatabaseFile())
	var source_occur int = 0
	var backup_occur int = 0

	if err != nil {
		return source_occur, backup_occur
	}

	stmt := `SELECT (
		select count(*) from source s where s.archive_id = ?
		) as source_occur, 
		(select count(*) from backup b where b.archive_id = ?
		) as backup_occur`

	row := conn.db.QueryRow(stmt, archive_id, archive_id)

	err = row.Scan(&source_occur, &backup_occur)

	if err != nil {
		fmt.Println("Cant get data from DB.")
	}

	return source_occur, backup_occur
}

func (conn *SQLite) AddBackupTimestamp(source_id int64, drive_ksuid string) int {
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return 0
	}

	var cnt int = 0

	stmt := `SELECT count(*) as cnt FROM timestamp WHERE source_id = ? AND drive_ksuid = ?`

	row := conn.db.QueryRow(stmt, source_id, drive_ksuid)

	err = row.Scan(&cnt)

	if err == nil && cnt != 0 {
		_, err := conn.Exec(`UPDATE timestamp SET archived_at = strftime('%s', 'now') where source_id = ? AND drive_ksuid = ?`, source_id, drive_ksuid)

		if err == nil {
			return 1
		}

		return 0
	}

	_, err = conn.Exec(`INSERT INTO timestamp(source_id, drive_ksuid) VALUES (?, ?)`, source_id, drive_ksuid)

	if err != nil {
		return 0
	}

	return 1
}

func (conn *SQLite) GetArchivesWithoutBackup() []int64 {
	var archive_ids = []int64{}
	var archive_id int64
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return archive_ids
	}

	stmt := `SELECT a.id from archive a 
	where (
	NOT EXISTS (select * from "source" s where s.archive_id = a.id)
	OR 
	NOT EXISTS (select * from "backup" b where b.archive_id = a.id)
	)`

	rows, err := conn.QueryRows(stmt)

	if err != nil {
		fmt.Println(err)
	}

	for rows.Next() {
		err := rows.Scan(&archive_id)

		if err != nil {
			fmt.Println(err)
			continue
		}

		archive_ids = append(archive_ids, archive_id)
	}

	return archive_ids
}

func (conn *SQLite) GetSourcesWithoutBackup() []int64 {
	var source_ids = []int64{}
	var source_id int64
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return source_ids
	}

	stmt := `select s.id from source s left join backup b on s.archive_id = b.archive_id WHERE b.archive_id IS NULL`
	rows, err := conn.QueryRows(stmt)

	if err != nil {
		fmt.Println(err)
	}

	for rows.Next() {
		err := rows.Scan(&source_id)

		if err != nil {
			fmt.Println(err)
			continue
		}

		source_ids = append(source_ids, source_id)
	}

	return source_ids
}

func (conn *SQLite) GetArchives() []Archive {
	var archives = []Archive{}
	var id int64
	var name string

	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return archives
	}

	stmt := `select id, name from archive`
	rows, err := conn.QueryRows(stmt)

	if err != nil {
		fmt.Println(err)
	}

	for rows.Next() {
		err := rows.Scan(&id, &name)

		if err != nil {
			fmt.Println(err)
			continue
		}

		archives = append(archives, Archive{id, name})
	}

	return archives
}

func (conn *SQLite) GetNewestTimestamp(database_path string) sql.NullTime {
	err := conn.OpenConnection(database_path)

	if err != nil {
		return sql.NullTime{}
	}

	var newest_timestamp = sql.NullTime{}

	stmt := `SELECT archived_at FROM timestamp ORDER BY archived_at DESC`

	row := conn.db.QueryRow(stmt)

	_ = row.Scan(&newest_timestamp)

	return newest_timestamp
}

func GenConditionSQL(source_ids []int64, archive_names []string) ([]interface{}, string) {
	var pos int = 0
	var sql_str string = " AND ("
	args := make([]interface{}, len(source_ids)+len(archive_names))

	for _, value := range source_ids {
		args[pos] = value
		sql_str += " s.id = ?"
		if pos < len(source_ids)-1 || len(archive_names) > 0 {
			sql_str += " OR"
		}
		pos++
	}

	for i, value := range archive_names {
		args[pos+i] = value
		sql_str += " a.name = ?"
		if i < len(source_ids)-1 {
			sql_str += " OR"
		}
	}

	sql_str += ")"

	return args, sql_str
}

func (conn *SQLite) FindBackups(source_ids []int64, archive_names []string) (map[int64]BackupRel, error) {
	backup_rels := make(map[int64]BackupRel)
	err := conn.OpenConnection(helper.GetDatabaseFile())

	var source_ksuid string
	var source_path sql.NullString
	var backup_ksuid string
	var backup_path sql.NullString
	var archive_name string
	var archive_id int64
	var archived_at sql.NullTime
	var source_id int64

	if err != nil {
		return backup_rels, err
	}

	sql_str := `SELECT 
	s.id as source_id,	
	s.drive_ksuid  as source_ksuid,
	s."path" as source_path,
	b.drive_ksuid as backup_ksuid,
	b."path" as backup_path,
	a.id as archive_id,
	a.name as archive_name,
	t.archived_at as archived_at
	from "backup" b 
	---------------
	left join "source" s 
	--------------------
	ON b.archive_id = s.archive_id 
	------------------------------
	left join archive a on b.archive_id = a.id 
	------------------------------------------------------
	left join drive s_d ON s.drive_ksuid = s_d.drive_ksuid 
	------------------------------------------------------
	left join drive b_d ON b.drive_ksuid = b_d.drive_ksuid
	left join "timestamp" t ON s.id = t.source_id and b.drive_ksuid = t.drive_ksuid 
	WHERE 1=1`

	var rows rows.Rows
	if len(source_ids) != 0 || len(archive_names) != 0 {
		args, sql_str_conditions := GenConditionSQL(source_ids, archive_names)

		rows, err = conn.QueryRows(sql_str+sql_str_conditions, args...)
	} else {
		rows, err = conn.QueryRows(sql_str)
	}

	if err != nil {
		return backup_rels, err
	}

	for rows.Next() {
		err := rows.Scan(&source_id, &source_ksuid, &source_path, &backup_ksuid, &backup_path, &archive_id, &archive_name, &archived_at)

		if err != nil {
			return backup_rels, err
		}

		backup := Backup{Ksuid: backup_ksuid, Path: backup_path}

		_, ok := backup_rels[source_id]

		if !ok {
			source := Source{Id: source_id, Ksuid: source_ksuid, Path: source_path}
			archive := Archive{Id: archive_id, Name: archive_name}

			backup_rels[source_id] = BackupRel{
				Source:      source,
				Backup:      []Backup{},
				Archive:     archive,
				Archived_at: archived_at,
			}
		}

		backup_rel := backup_rels[source_id]
		backup_rel.Backup = append(backup_rel.Backup, backup)

		backup_rels[source_id] = backup_rel
	}

	rows.Close()

	return backup_rels, nil
}

func (conn *SQLite) ArchiveExists(archive_name string) bool {
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return false
	}

	var cnt int = 0

	stmt := `SELECT count(*) as cnt FROM archive WHERE name = ?`

	row := conn.db.QueryRow(stmt, archive_name)

	err = row.Scan(&cnt)

	if err == nil && cnt != 0 {
		return true
	}

	return false
}

func (conn *SQLite) GetArchiveID(archive_name string) int64 {
	var archive_id int64 = 0
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return archive_id
	}

	stmt := `SELECT id FROM archive WHERE name = ?`

	row := conn.db.QueryRow(stmt, archive_name)

	err = row.Scan(&archive_id)

	if err != nil {
		fmt.Println(err)
	}

	return archive_id
}

func (conn *SQLite) CreateArchive(archive_name string) (int64, error) {

	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return 0, err
	}

	stmt := `SELECT id FROM archive WHERE name = ?`
	var archive_id int64

	row := conn.db.QueryRow(stmt, archive_name)
	err = row.Scan(&archive_id)

	if err == nil {
		return archive_id, errors.New("Archive name is not unique.")
	}

	result, err := conn.Exec(`INSERT INTO archive(name) VALUES (?)`, archive_name)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	return id, err
}

func (conn *SQLite) CreateSource(drive_ksuid string, path string) int64 {

	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return 0
	}

	var id int64

	stmt := `SELECT id FROM source WHERE drive_ksuid = ? AND path = ?`

	row := conn.db.QueryRow(stmt, drive_ksuid, path)
	err = row.Scan(&id)

	if err == nil {
		fmt.Println("Source path already exists, archive will be updated.")
		return -1
	}

	result, err := conn.Exec(`INSERT INTO source(drive_ksuid, path) VALUES (?, ?)`, drive_ksuid, path)

	if err != nil {
		fmt.Println(err)
		return 0
	}

	id, _ = result.LastInsertId()

	return id
}

func (conn *SQLite) ClearAllTables() error {

	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return err
	}

	tables := []string{"backup", "source", "archive", "drive", "timestamp"}

	for _, table := range tables {
		_, err = conn.db.Exec(`DELETE FROM ` + table)

		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	return nil
}

func (conn *SQLite) CreateBackup(archive_id int64, drive_ksuid string, path string) error {

	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return err
	}

	stmt := `SELECT archive_id, drive_ksuid FROM backup WHERE archive_id = ? AND drive_ksuid = ? AND path = ?`

	row := conn.db.QueryRow(stmt, archive_id, drive_ksuid, path)
	err = row.Scan(&archive_id, &drive_ksuid)

	if err == nil {
		return err
	}

	_, err = conn.Exec(`INSERT INTO backup(archive_id, drive_ksuid, path) VALUES (?, ?, ?)`, archive_id, drive_ksuid, path)

	if err == nil {
		return err
	}

	return nil
}

func (conn *SQLite) UpdateSourceArchive(source_id int64, archive_id int64) error {

	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return err
	}

	_, err = conn.Exec(`UPDATE source SET archive_id = ? where id = ?`, archive_id, source_id)

	if err != nil {
		return err
	}

	return nil
}

func (conn *SQLite) UpdateDriveName(drive_ksuid string, new_name string) int {

	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return 0
	}

	_, err = conn.Exec(`UPDATE drive SET name = ? where drive_ksuid = ?`, new_name, drive_ksuid)

	if err != nil {
		return 0
	}

	return 1
}

// Creates database if doesnt exist
func CreateDB() error {
	var database_path = helper.GetDatabaseFile()

	if helper.Exists(database_path) != nil {
		err := os.MkdirAll(filepath.Dir(database_path), os.ModePerm)

		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		file, err := os.Create(database_path)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		file.Close()
		fmt.Println("sqlite-database.db created")

	}

	return nil
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
			name VARCHAR NOT NULL UNIQUE
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
			archived_at timestamp DEFAULT (strftime('%s', 'now')) NOT NULL,
			CONSTRAINT timestamp_PK PRIMARY KEY (source_id,drive_ksuid),
			CONSTRAINT timestamp_FK FOREIGN KEY (source_id) REFERENCES "source"(id),
			CONSTRAINT timestamp_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid)
		);`,
	}

	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return err
	}

	for _, table := range tables {
		_, err := conn.Exec(table)

		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}

func (conn *SQLite) OpenConnection(database_path string) error {
	var err error
	conn.db, err = sql.Open("sqlite3", database_path) // Open the created SQLite File

	if err != nil {
		fmt.Println(err)
	}

	return conn.db.Ping()
}

func (conn *SQLite) TestDatabase(database_path string) map[int64]Timestamp {
	var records_map = map[int64]Timestamp{}

	if helper.Exists(database_path) != nil {
		return records_map
	}

	err := conn.OpenConnection(database_path)

	if err != nil {
		return records_map
	}

	stmt := `SELECT source_id, drive_ksuid, archived_at FROM timestamp`

	rows, err := conn.QueryRows(stmt)

	if err != nil {
		return records_map
	}

	conn.db.Close()

	for rows.Next() {
		var record Timestamp

		err := rows.Scan(&record.Source_id, &record.Drive_ksuid, &record.Archived_at)

		if err != nil {
			continue
		}

		records_map[record.Source_id] = record

		//records = append(records, record)
	}

	rows.Close()

	return records_map

}

// Inserts record into table drives
func (conn *SQLite) InsertDriveDB(ksuid string, drive_name string) int64 {

	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return 0
	}

	var result sql.Result

	if drive_name == "" {
		result, err = conn.Exec(`INSERT INTO drive(drive_ksuid) VALUES (?)`, ksuid)
	} else {
		result, err = conn.Exec(`INSERT INTO drive(drive_ksuid, name) VALUES (?, ?)`, ksuid, drive_name)
	}

	if err != nil {
		return 0
	}

	id, err := result.LastInsertId()

	return id
}

// Returns drive by ksuid from db
func (conn *SQLite) DriveInDB(drive_ksuid string) (string, sql.NullString) {
	err := conn.OpenConnection(helper.GetDatabaseFile())

	if err != nil {
		return "", sql.NullString{}
	}

	stmt := `SELECT drive_ksuid, name FROM drive
	WHERE drive_ksuid=?`
	var r_ksuid string
	var r_name sql.NullString

	row := conn.db.QueryRow(stmt, drive_ksuid)
	switch err := row.Scan(&r_ksuid, &r_name); err {
	case nil:
		return r_ksuid, r_name
	default:
		return "", sql.NullString{}
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
