ALTER TABLE tracker_queries DROP CONSTRAINT tracker_queries_tracker_id_trackers_id_foreign;
ALTER TABLE tracker_items DROP CONSTRAINT tracker_items_tracker_id_trackers_id_foreign;
ALTER TABLE trackers DROP CONSTRAINT trackers_pkey CASCADE;
ALTER TABLE tracker_queries DROP COLUMN id;
ALTER TABLE trackers ADD COLUMN "ID" UUID primary key DEFAULT uuid_generate_v4() NOT NULL;
CREATE UNIQUE INDEX trackers_idx ON trackers USING btree ("ID");
ALTER TABLE tracker_queries
	ADD CONSTRAINT tracker_queries_trackers_id_trackers_ID_foreign
		FOREIGN KEY (tracker_id)
		REFERENCES trackers(ID);
