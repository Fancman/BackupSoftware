package database

type Source struct {
	Id    int64
	Ksuid string
	Path  string
}

type Backup struct {
	Ksuid string
	Path  string
}

type Archive struct {
	Id   int64
	Name string
}

type BackupRel struct {
	Source  Source
	Backup  []Backup
	Archive Archive
}
