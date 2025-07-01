DO $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'vehicle_movement' 
        AND column_name = 'sent_unix_time'
    ) THEN
        IF NOT EXISTS (
            SELECT 1 
            FROM information_schema.columns 
            WHERE table_name = 'vehicle_movement' 
            AND column_name = 'sent_at'
        ) THEN
            ALTER TABLE vehicle_movement ADD COLUMN sent_at TIMESTAMP;
        END IF;
        
        IF NOT EXISTS (
            SELECT 1 
            FROM information_schema.columns 
            WHERE table_name = 'vehicle_movement' 
            AND column_name = 'received_at'
        ) THEN
            ALTER TABLE vehicle_movement ADD COLUMN received_at TIMESTAMP;
        END IF;

        UPDATE vehicle_movement
        SET 
            sent_at = TO_TIMESTAMP(sent_unix_time),
            received_at = TO_TIMESTAMP(received_unix_time)
        WHERE sent_at IS NULL OR received_at IS NULL;

        ALTER TABLE vehicle_movement DROP COLUMN sent_unix_time;
        ALTER TABLE vehicle_movement DROP COLUMN received_unix_time;
    END IF;
END $$;

ALTER TABLE vehicle_movement ALTER COLUMN received_at SET NOT NULL;