-- Write your migrate up statements here
create table GuildCount(
    id serial primary key,
    guild_count integer,
    timestamp integer
);

---- create above / drop below ----
drop table GuildCount;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
