BEGIN;

alter table guildcount
    add column created_at timestamp without time zone;

alter table guildcount
    alter column created_at set default now() at time zone ('utc');

COMMIT;