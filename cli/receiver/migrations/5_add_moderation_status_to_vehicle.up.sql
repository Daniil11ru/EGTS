CREATE TYPE moderation_status AS ENUM (
  'pending',
  'approved',
  'rejected'
);

ALTER TABLE vehicle
  ADD COLUMN moderation_status moderation_status NOT NULL
    DEFAULT 'approved';