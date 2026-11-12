-- 000012_populate_clinic_coordinates.up.sql
-- Set exact verified geographic coordinates for existing approved clinics in Termez

-- 1. AURAMED 24/7 (Islom Karimov ko'chasi 177-b)
UPDATE clinics
SET latitude = 37.234278,
    longitude = 67.292365,
    updated_at = CURRENT_TIMESTAMP
WHERE id = 'fa1edf59-8f15-499f-8817-0f109f771f7f'
  AND (latitude IS NULL OR longitude IS NULL);

-- 2. Sevinch parazitoligiya klinikasi (Jurabayeva ko'chasi)
UPDATE clinics
SET latitude = 37.244245,
    longitude = 67.277320,
    updated_at = CURRENT_TIMESTAMP
WHERE id = '201f09c6-2ada-4150-953b-9815f0fa6a04'
  AND (latitude IS NULL OR longitude IS NULL);

-- 3. Stomatologiya (Pyati endokrinologiya ro'parasi, Serenada to'yxonasi)
UPDATE clinics
SET latitude = 37.239278,
    longitude = 67.293741,
    updated_at = CURRENT_TIMESTAMP
WHERE id = '6dfcf905-1a14-412d-b6de-ee0d9645e0ca'
  AND (latitude IS NULL OR longitude IS NULL);

