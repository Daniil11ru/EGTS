ALTER TABLE vehicle_movement
  DROP COLUMN oid,
  DROP COLUMN latitude,
  DROP COLUMN longitude,
  DROP COLUMN altitude,
  DROP COLUMN direction,
  DROP COLUMN speed,
  DROP COLUMN navigation_system,
  DROP COLUMN satellite_count,
  DROP COLUMN sent_unix_time,
  DROP COLUMN received_unix_time;

ALTER TABLE vehicle_movement ADD COLUMN data JSONB;