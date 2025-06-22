ALTER TABLE public.vehicle_directory
    DROP COLUMN IF EXISTS imei_extraction_unit_type,
    DROP COLUMN IF EXISTS imei_extraction_position_type,
    DROP COLUMN IF EXISTS imei_segment_length;
