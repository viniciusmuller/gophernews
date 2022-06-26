alter table users
alter column email type VARCHAR(120),
drop column creation_date timestamptz,
drop column last_modification_date timestamptz;
