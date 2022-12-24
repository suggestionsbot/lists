BEGIN;

alter table guildcount
    drop column created_at;

COMMIT;