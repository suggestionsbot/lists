BEGIN;

alter table guildcount
    add column shard_count integer;

COMMIT;