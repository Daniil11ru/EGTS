BEGIN;

CREATE TABLE IF NOT EXISTS provider_to_ip (
    provider_id INTEGER NOT NULL REFERENCES provider(id),
    ip varchar(15) NOT NULL
);

INSERT INTO provider_to_ip (provider_id, ip)
SELECT id, ip FROM provider WHERE ip IS NOT NULL;

ALTER TABLE provider DROP COLUMN IF EXISTS ip;

COMMIT;