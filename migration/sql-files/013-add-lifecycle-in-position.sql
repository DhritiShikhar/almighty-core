ALTER TABLE position 
    ADD COLUMN created_at timestamp with time zone,
    ADD COLUMN updated_at timestamp with time zone,
    ADD COLUMN deleted_at timestamp with time zone;
	
