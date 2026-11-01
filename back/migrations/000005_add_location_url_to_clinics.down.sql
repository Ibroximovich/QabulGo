-- 000005_add_location_url_to_clinics.down.sql
ALTER TABLE clinics DROP COLUMN IF EXISTS location_url;

