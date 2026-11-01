-- 000005_add_location_url_to_clinics.up.sql
ALTER TABLE clinics ADD COLUMN IF NOT EXISTS location_url VARCHAR(500);

