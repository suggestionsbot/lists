BEGIN;

alter table guildcount
    drop column shard_count;

COMMIT;