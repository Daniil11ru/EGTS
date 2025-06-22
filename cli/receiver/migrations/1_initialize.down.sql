DROP TABLE IF EXISTS public.vehicle_movement;
DROP SEQUENCE IF EXISTS telematics_subrecord_id_seq;

DROP TABLE IF EXISTS public.vehicle;
DROP SEQUENCE IF EXISTS vehicle_id_seq;

DROP TABLE IF EXISTS public.vehicle_directory;
DROP SEQUENCE IF EXISTS vehicle_directory_metadata_id_seq;
DROP TYPE IF EXISTS public.imei_extraction_position_type;
DROP TYPE IF EXISTS public.imei_extraction_unit_type;

DROP TABLE IF EXISTS public.provider_to_ip;

DROP TABLE IF EXISTS public.provider;
DROP SEQUENCE IF EXISTS provider_id_seq;
