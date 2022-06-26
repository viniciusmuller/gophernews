alter table users
alter column email type VARCHAR(254),
add creation_date timestamptz default now(),
add last_modification_date timestamptz default now();
