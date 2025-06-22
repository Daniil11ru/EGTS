ALTER TABLE public.vehicle_directory
    ADD COLUMN IF NOT EXISTS imei_extraction_unit_type public.imei_extraction_unit_type NOT NULL,
    ADD COLUMN IF NOT EXISTS imei_extraction_position_type public.imei_extraction_position_type NOT NULL,
    ADD COLUMN IF NOT EXISTS imei_segment_length int2;