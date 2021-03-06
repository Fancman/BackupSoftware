package database

import "database/sql"

type Source struct {
	Id    int64
	Ksuid string
	Path  sql.NullString
}

type Backup struct {
	Ksuid string
	Path  sql.NullString
}

type Archive struct {
	Id   int64
	Name string
}

type BackupRel struct {
	Source      Source
	Backup      []Backup
	Archive     Archive
	Archived_at sql.NullTime
}

type BackupPaths struct {
	Sources     []string
	SourceIDs   []int64
	Destination string
	BackupKsuid string
}

type Timestamp struct {
	Source_id   int64
	Drive_ksuid string
	Archived_at sql.NullTime
}

type DriveRecord struct {
	Letter         string
	Name           string
	File_exists    bool
	File_accesible bool
	Ksuid          string
	Timestamp      string
}
