BEGIN;

ALTER TABLE location ADD COLUMN oid BIGINT;

UPDATE location l
SET oid = v.oid
FROM vehicle v
WHERE v.id = l.vehicle_id;

ALTER TABLE location ALTER COLUMN oid SET NOT NULL;

COMMIT;
