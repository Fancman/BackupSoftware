CREATE TABLE drive (
	drive_ksuid INTEGER NOT NULL,
	name VARCHAR,
	CONSTRAINT drive_PK PRIMARY KEY (drive_ksuid)
);
CREATE INDEX drive_drive_ksuid_IDX ON drive (drive_ksuid);


CREATE TABLE archive (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name VARCHAR
);

CREATE TABLE backup (
	archive_id INTEGER,
	drive_ksuid INTEGER,
	"path" VARCHAR,
	CONSTRAINT backup_PK PRIMARY KEY (archive_id,drive_ksuid),
	CONSTRAINT backup_FK FOREIGN KEY (archive_id) REFERENCES archive(id) ON DELETE RESTRICT ON UPDATE CASCADE,
	CONSTRAINT backup_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid) ON DELETE RESTRICT ON UPDATE CASCADE
);

CREATE TABLE "source" (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	archive_id INTEGER,
	drive_ksuid INTEGER,
	"path" VARCHAR,
	CONSTRAINT source_FK FOREIGN KEY (id) REFERENCES archive(id) ON DELETE RESTRICT ON UPDATE CASCADE,
	CONSTRAINT source_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid) ON DELETE RESTRICT ON UPDATE CASCADE
);

CREATE TABLE "timestamp" (
	source_id INTEGER,
	drive_ksuid INTEGER,
	created_at timestamp DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT timestamp_PK PRIMARY KEY (source_id,drive_ksuid),
	CONSTRAINT timestamp_FK FOREIGN KEY (source_id) REFERENCES "source"(id),
	CONSTRAINT timestamp_FK_1 FOREIGN KEY (drive_ksuid) REFERENCES drive(drive_ksuid)
);
