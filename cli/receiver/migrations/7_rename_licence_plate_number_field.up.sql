BEGIN;

ALTER TABLE vehicle RENAME COLUMN license_plate_number TO name;

COMMIT;