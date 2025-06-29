BEGIN;

ALTER TABLE vehicle RENAME COLUMN name TO license_plate_number;

COMMIT;