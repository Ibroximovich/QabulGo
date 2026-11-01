-- 000006_add_geolocation_to_clinics.up.sql

ALTER TABLE clinics
    ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION;

-- Check constraint: either both NULL or both within valid ranges
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_clinic_coordinates'
    ) THEN
        ALTER TABLE clinics
            ADD CONSTRAINT chk_clinic_coordinates
            CHECK (
                (latitude IS NULL AND longitude IS NULL) OR
                (latitude IS NOT NULL AND longitude IS NOT NULL AND latitude BETWEEN -90.0 AND 90.0 AND longitude BETWEEN -180.0 AND 180.0)
            );
    END IF;
END $$;

-- Spatial search index
CREATE INDEX IF NOT EXISTS idx_clinics_geolocation
    ON clinics(latitude, longitude)
    WHERE latitude IS NOT NULL AND longitude IS NOT NULL;

