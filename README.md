# Bakalarska praca

Nástroj na distribuované zálohovanie

[![Build status](https://ci.appveyor.com/api/projects/status/github/Fancman/BackupSoftware?svg=TRUE)](https://ci.appveyor.com/project/Fancman/BackupSoftware)



### How to get it

To get the latest version clone project to current directory and build it from source:

```
git clone https://github.com/Fancman/BackupSoftware
cd BackupSoftware
go build .
```

Program should be now succesfully built. You can also use already built project from release page on [github.com](https://github.com/Fancman/BackupSoftware/releases) or download it from [AppVeyor](https://ci.appveyor.com/project/Fancman/backupsoftware) from subpage "Artifacts".



### Running app
List all commands with command  `help`


Usage:
  Backupsoft.exe [command] 

Available Commands:
  - add-drive -- Adds drive to db with optional custom name and creates .drive file
	```
	Backupsoft.exe add-drive [drive letter] -n [Custom drive name]

	Flags:
	-n, --drive-name string   Drive name
	-h, --help                help for add-drive
	```

	Usage:

	```
	Backupsoft.exe add-drive E -n "E drive"

	```

  - clear-tables -- Deletes all records from tables.
  	```
	Backupsoft.exe clear-tables
	```

  - create-backup -- Create backup record from source and destination paths. Archive name is optional.
 	```
	Backupsoft.exe create-backup -s [source paths] -d [destination paths] -a [archive name]

	Flags:
	-a, --archive string            archive name
	-d, --destination stringArray   destination path
	-h, --help                      help for create-backup
	-s, --source stringArray        sources paths
	```

	Usage:

	```
	Backupsoft.exe create-backup -s "E:\test" -s "C:\DEV\web\floorplan" -d "D:\backup" -a "secret-files.7z"

	```

  - list-backups -- List stored backup records
	```
	Backupsoft.exe list-backups
	```

  - list-drives -- List available drives.
  	```
	Backupsoft.exe list-drives
	```

  - load-db --Load database from drive
  	```
	Backupsoft.exe load-db -d [drive letter]

	Flags:
	-l, --drive-letter string   Drive letter
	-h, --help                  help for load-db
	```

	Usage:

	```
	Backupsoft.exe load-db "E"

	```

  - remove-backup -- Remove backup records by archive_id and drive_ksuid, path to destination, archive name or drive letter
  
	```
	Backupsoft.exe remove-backup -i [archive id] -d [drive ksuid] | -p [path to destination] | -a [archive name] | -l [drive letter]

	Flags:
	-i, --archive-id int        Archive ID
	-a, --archive-name string   Archive name
	-p, --dest-path string      Destination path
	-d, --drive-ksuid string    Drive Ksuid
	-l, --drive-letter string   Drive letter
	-h, --help                  help for remove-backup
	```

	Usage:

	```
	Backupsoft.exe remove-backup -i 10 -d "1mC60uVtv07vvPY4ylFkaXlc4b9"

	Backupsoft.exe remove-backup -p "D:\backup\archive-name.7z"

	Backupsoft.exe remove-backup -a "archive-name.7z"

	Backupsoft.exe remove-backup -l "E"
	```

  - remove-source -- Remove source by source ids
	```
	Backupsoft.exe remove-source -s [source id]

	Flags:
	-h, --help            help for remove-source
	-s, --source-id int   Source ID
	```

	Usage:

	```
	Backupsoft.exe remove-source -s 10
	```

  - start-backup  -- Start backup for record in db by its ids
  	```
	Backupsoft.exe start-backup -s [source ids] -a [archive name]

	Flags:
	-a, --archive stringArray   archive name
	-h, --help                  help for start-backup
	-s, --source int64Slice     source ids (default [])
	```

	Usage:

	```
	Backupsoft.exe start-backup -s 10

	Backupsoft.exe start-backup -a "archive-name.7z"
	```

  - start-restore -- Start restore from record in db by its ids or archive names, optional backup paths (extract only this archive)
	```
	Backupsoft.exe start-restore -s [source ids] | -a [archive names] -b [backup paths]

	Flags:
	-a, --archive stringArray   archive name
	-b, --backup stringArray    backup paths
	-h, --help                  help for start-restore
	-s, --source int64Slice     source ids (default [])	
	```	

	Usage:

	```
	Backupsoft.exe start-restore -s 10 -s 11

	Backupsoft.exe start-restore -s 10 -b "D:\backup\archive-name.7z"

	Backupsoft.exe start-restore -a "archive-name.7z" -b "D:\backup\archive-name.7z"
	```

  - help -- Help about any command