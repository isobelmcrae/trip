-- sqlite3 api.sqlite
create table "stop" (
	"id" text not null primary key,
	"name" text not null,
	"lat" real not null,
	"lon" real not null
);

create virtual table if not exists "stop_fts" using fts5(
    id unindexed,
    name,
    content='',
    contentless_unindexed=1
);
