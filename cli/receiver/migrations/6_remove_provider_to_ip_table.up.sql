BEGIN;

ALTER TABLE provider ADD COLUMN ip TEXT;

UPDATE provider p
SET ip = sub.ip
FROM (
    SELECT DISTINCT ON (provider_id) provider_id, ip
    FROM provider_to_ip
    ORDER BY provider_id, ip
) AS sub
WHERE p.id = sub.provider_id;

DROP TABLE IF EXISTS provider_to_ip;

COMMIT;