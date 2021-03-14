# Bakalarska praca

Nástroj na distribuované zálohovanie

[![Build status](https://ci.appveyor.com/api/projects/status/github/Fancman/BackupSoftware?svg=TRUE)](https://ci.appveyor.com/project/Fancman/BackupSoftware)


## Navod

- Listnutie dostupnych drivov: list-drives
- Pridanie drivu do dtabazy a vytvorenie .drive subora: add-drive [drive letter] 
- Vytvorenie zaznamu zalohy: create-backup -s [source path] -d [destination path] -p [custom path on drive]
- Listnutie zaznamov zaloh: list-backups
- Spustenie zalohovania pre zaznam: start-backup [backup id]
- Obnovenie podla zaznamu zalohy: start-restore [backup id]