-- 000014_add_new_doctor_specialties.down.sql
-- Yangi 21 ta shifokor mutaxassisliklarini qaytarish (revert)

DELETE FROM specialties WHERE slug IN (
    'onkolog',
    'gematolog',
    'pulmonolog',
    'nefrolog',
    'revmatolog',
    'infeksionist',
    'psixiatr',
    'psixoterapevt',
    'neyroxirurg',
    'anesteziolog-reanimatolog',
    'mammolog',
    'proktolog',
    'androlog',
    'gepatolog',
    'angioxirurg',
    'onkoginekolog',
    'bolalar-nevrologi',
    'bolalar-jarrohi',
    'neonatolog',
    'immunolog',
    'parazitolog'
);

