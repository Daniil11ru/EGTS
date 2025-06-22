DROP TABLE IF EXISTS "public"."provider";
-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS provider_id_seq;

-- Table Definition
CREATE TABLE "public"."provider" (
    "id" int4 NOT NULL DEFAULT nextval('provider_id_seq'::regclass),
    "name" varchar(255) NOT NULL,
    PRIMARY KEY ("id")
);

DROP TABLE IF EXISTS "public"."provider_to_ip";
-- Table Definition
CREATE TABLE "public"."provider_to_ip" (
    "provider_id" int4 NOT NULL,
    "ip" varchar(15) NOT NULL,
    CONSTRAINT "provider_to_ip_provider_id_fkey" FOREIGN KEY ("provider_id") REFERENCES "public"."provider"("id") ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY ("provider_id","ip")
);

DROP TABLE IF EXISTS "public"."vehicle";
-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS vehicle_id_seq;

DROP TABLE IF EXISTS "public"."vehicle_directory";
-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS vehicle_directory_metadata_id_seq;
DROP TYPE IF EXISTS "public"."imei_extraction_unit_type";
CREATE TYPE "public"."imei_extraction_unit_type" AS ENUM ('digits', 'bytes');
DROP TYPE IF EXISTS "public"."imei_extraction_position_type";
CREATE TYPE "public"."imei_extraction_position_type" AS ENUM ('prefix', 'suffix');

-- Table Definition
CREATE TABLE "public"."vehicle_directory" (
    "id" int4 NOT NULL DEFAULT nextval('vehicle_directory_metadata_id_seq'::regclass),
    "provider_id" int4 NOT NULL,
    "imei_extraction_unit_type" "public"."imei_extraction_unit_type" NOT NULL,
    "imei_extraction_position_type" "public"."imei_extraction_position_type" NOT NULL,
    "imei_segment_length" int2,
    CONSTRAINT "vehicle_directory_metadata_provider_id_fkey" FOREIGN KEY ("provider_id") REFERENCES "public"."provider"("id") ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY ("id")
);

-- Table Definition
CREATE TABLE "public"."vehicle" (
    "imei" int8 NOT NULL,
    "license_plate_number" varchar(8),
    "vehicle_directory_id" int4 NOT NULL,
    "id" int4 NOT NULL DEFAULT nextval('vehicle_id_seq'::regclass),
    CONSTRAINT "alllowed_vehicle_vehicle_directory_metadata_id_fkey" FOREIGN KEY ("vehicle_directory_id") REFERENCES "public"."vehicle_directory"("id") ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY ("id")
);

-- Indices
CREATE UNIQUE INDEX vehicle_directory_metadata_pkey ON public.vehicle_directory USING btree (id);

DROP TABLE IF EXISTS "public"."vehicle_movement";
-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS telematics_subrecord_id_seq;

-- Table Definition
CREATE TABLE "public"."vehicle_movement" (
    "id" int4 NOT NULL DEFAULT nextval('telematics_subrecord_id_seq'::regclass),
    "data" jsonb NOT NULL,
    "vehicle_id" int4 NOT NULL,
    CONSTRAINT "vehicle_movement_vehicle_id_fkey" FOREIGN KEY ("vehicle_id") REFERENCES "public"."vehicle"("id") ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY ("id")
);


-- Indices
CREATE UNIQUE INDEX telematics_subrecord_pkey ON public.vehicle_movement USING btree (id);

INSERT INTO "public"."provider" ("id", "name") VALUES
(1, 'Локальный');
INSERT INTO "public"."provider_to_ip" ("provider_id", "ip") VALUES
(1, '127.0.0.1');
