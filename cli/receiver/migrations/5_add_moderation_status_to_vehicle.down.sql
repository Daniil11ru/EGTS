ALTER TABLE vehicle
  DROP COLUMN IF EXISTS moderation_status;

DROP TYPE IF EXISTS moderation_status;
