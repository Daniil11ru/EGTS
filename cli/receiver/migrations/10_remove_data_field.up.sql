DELETE FROM vehicle_movement;

ALTER TABLE vehicle_movement DROP COLUMN data;

ALTER TABLE vehicle_movement
  ADD COLUMN oid BIGINT NOT NULL,
  ADD COLUMN latitude SMALLINT,
  ADD COLUMN longitude SMALLINT,
  ADD COLUMN altitude SMALLINT,
  ADD COLUMN direction SMALLINT,
  ADD COLUMN speed SMALLINT,
  ADD COLUMN navigation_system SMALLINT,
  ADD COLUMN satellite_count SMALLINT,
  ADD COLUMN sent_unix_time BIGINT,
  ADD COLUMN received_unix_time BIGINT NOT NULL;