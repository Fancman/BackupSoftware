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
