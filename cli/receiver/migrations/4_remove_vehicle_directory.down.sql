CREATE SEQUENCE IF NOT EXISTS vehicle_directory_metadata_id_seq;

CREATE TABLE public.vehicle_directory (
    id int4 NOT NULL DEFAULT nextval('vehicle_directory_metadata_id_seq'::regclass),
    provider_id int4 NOT NULL,
    PRIMARY KEY (id)
);

ALTER TABLE public.vehicle_directory
ADD CONSTRAINT vehicle_directory_metadata_provider_id_fkey
FOREIGN KEY (provider_id) REFERENCES public.provider(id)
ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE public.vehicle ADD COLUMN vehicle_directory_id int4;

INSERT INTO public.vehicle_directory (provider_id)
SELECT DISTINCT provider_id FROM public.vehicle;

UPDATE public.vehicle v
SET    vehicle_directory_id = vd.id
FROM   public.vehicle_directory vd
WHERE  vd.provider_id = v.provider_id;

ALTER TABLE public.vehicle ALTER COLUMN vehicle_directory_id SET NOT NULL;

ALTER TABLE public.vehicle
ADD CONSTRAINT allowed_vehicle_vehicle_directory_metadata_id_fkey
FOREIGN KEY (vehicle_directory_id) REFERENCES public.vehicle_directory(id)
ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE public.vehicle DROP CONSTRAINT vehicle_provider_id_fkey;

ALTER TABLE public.vehicle DROP COLUMN provider_id;
