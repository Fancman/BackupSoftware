# Bakalarska praca

Nástroj na distribuované zálohovanie

[![Build status](https://ci.appveyor.com/api/projects/status/github/Fancman/BackupSoftware?svg=TRUE)](https://ci.appveyor.com/project/Fancman/BackupSoftware)


## Navod

- Listnutie dostupnych drivov vo formate ([drive letter] - [ksuid ak ma] - [status]). Prikaz: list-drives
- Pridanie drivu do databazy a vytvorenie .drive subora. Prikaz: add-drive [drive letter] 
- Vytvorenie zaznamu zalohy. Prikaz: create-backup -s [source path] -d [destination drive ksuid] -p [custom path on drive]
- Listnutie zaznamov zaloh vo formate ([backup id] [source path] [destination drive ksuid]). Prikaz: list-backups
- Spustenie zalohovania pre zaznam. Prikaz: start-backup [backup id]
- Obnovenie podla zaznamu zalohy. Prikaz: start-restore [backup id]