DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'vehicle_movement' 
        AND column_name = 'sent_unix_time'
    ) THEN
        ALTER TABLE vehicle_movement ADD COLUMN sent_unix_time BIGINT;
        ALTER TABLE vehicle_movement ADD COLUMN received_unix_time BIGINT;
        
        UPDATE vehicle_movement
        SET 
            sent_unix_time = EXTRACT(EPOCH FROM sent_at)::BIGINT,
            received_unix_time = EXTRACT(EPOCH FROM received_at)::BIGINT;
            
        ALTER TABLE vehicle_movement DROP COLUMN sent_at;
        ALTER TABLE vehicle_movement DROP COLUMN received_at;
    END IF;
END $$;