ALTER TABLE public.vehicle ADD COLUMN provider_id int4;

UPDATE public.vehicle v
SET    provider_id = vd.provider_id
FROM   public.vehicle_directory vd
WHERE  vd.id = v.vehicle_directory_id;

ALTER TABLE public.vehicle ALTER COLUMN provider_id SET NOT NULL;

ALTER TABLE public.vehicle
ADD CONSTRAINT vehicle_provider_id_fkey
FOREIGN KEY (provider_id) REFERENCES public.provider(id)
ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE public.vehicle DROP CONSTRAINT allowed_vehicle_vehicle_directory_metadata_id_fkey;

ALTER TABLE public.vehicle DROP COLUMN vehicle_directory_id;

DROP TABLE public.vehicle_directory;

DROP SEQUENCE IF EXISTS vehicle_directory_metadata_id_seq;
