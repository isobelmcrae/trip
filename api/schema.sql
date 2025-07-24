-- sqlite3 api.sqlite
create table if not exists "stop" (
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

CREATE TABLE IF NOT EXISTS "stop_times" (
    "trip_id" TEXT NOT NULL,
    "stop_id" TEXT NOT NULL,
    "stop_sequence" INTEGER NOT NULL,
	"shape_dist_traveled" REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS "trips" (
    "trip_id" TEXT NOT NULL PRIMARY KEY,
    "route_id" TEXT NOT NULL,
    "shape_id" TEXT
);

CREATE TABLE IF NOT EXISTS "shapes" (
    "shape_id" TEXT NOT NULL,
    "shape_pt_lat" REAL NOT NULL,
    "shape_pt_lon" REAL NOT NULL,
    "shape_pt_sequence" INTEGER NOT NULL,
	"shape_dist_traveled" REAL NOT NULL,
    PRIMARY KEY (shape_id, shape_pt_sequence)
);

CREATE INDEX IF NOT EXISTS "idx_stop_times_trip_id" ON "stop_times" ("trip_id");
CREATE INDEX IF NOT EXISTS "idx_stop_times_stop_id" ON "stop_times" ("stop_id");
CREATE INDEX IF NOT EXISTS "idx_trips_shape_id" ON "trips" ("shape_id");
CREATE INDEX IF NOT EXISTS "idx_shapes_shape_id" ON "shapes" ("shape_id");
