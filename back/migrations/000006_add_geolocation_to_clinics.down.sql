-- 000006_add_geolocation_to_clinics.down.sql

DROP INDEX IF EXISTS idx_clinics_geolocation;
ALTER TABLE clinics DROP CONSTRAINT IF EXISTS chk_clinic_coordinates;
ALTER TABLE clinics DROP COLUMN IF EXISTS latitude;
ALTER TABLE clinics DROP COLUMN IF EXISTS longitude;

