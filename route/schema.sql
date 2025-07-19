-- sqlite3 route.sqlite
create table "stop" (
	"id" text not null primary key,
	"name" text not null,
	"lat" real not null,
	"lon" real not null
);

CREATE VIRTUAL TABLE IF NOT EXISTS "stop_fts" USING fts5(
    id UNINDEXED, -- This will be stored
    name,              -- This will be indexed but not stored
    content='',        -- Makes it contentless...
    contentless_unindexed=1 -- ...but tells it to store UNINDEXED columns.
);

/*

SELECT s.id, s.name, s.lat, s.lon
FROM stop_fts AS fts
JOIN stop AS s ON fts.id = s.id
WHERE fts.name MATCH 'Central Stat*';

*/