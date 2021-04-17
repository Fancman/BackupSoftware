SELECT b.archive_id, 
s.drive_ksuid  as source_ksuid,
s."path" as source_path,
s.id as source_id,
b.drive_ksuid as backup_ksuid,
b."path" as backup_path,
a.id as archive_id,
a.name as archive_name,
s_d.name as source_drive_name,
b_d.name as backup_drive_name 
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