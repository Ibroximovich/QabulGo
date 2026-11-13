-- 000014_add_new_doctor_specialties.up.sql
-- Yangi 21 ta shifokor mutaxassisliklarini qo'shish

INSERT INTO specialties (name, slug) VALUES
    ('Onkolog', 'onkolog'),
    ('Gematolog', 'gematolog'),
    ('Pulmonolog', 'pulmonolog'),
    ('Nefrolog', 'nefrolog'),
    ('Revmatolog', 'revmatolog'),
    ('Infeksionist', 'infeksionist'),
    ('Psixiatr', 'psixiatr'),
    ('Psixoterapevt', 'psixoterapevt'),
    ('Neyroxirurg', 'neyroxirurg'),
    ('Anesteziolog-reanimatolog', 'anesteziolog-reanimatolog'),
    ('Mammolog', 'mammolog'),
    ('Proktolog', 'proktolog'),
    ('Androlog', 'androlog'),
    ('Gepatolog', 'gepatolog'),
    ('Angioxirurg', 'angioxirurg'),
    ('Onkoginekolog', 'onkoginekolog'),
    ('Bolalar nevrologi', 'bolalar-nevrologi'),
    ('Bolalar jarrohi', 'bolalar-jarrohi'),
    ('Neonatolog', 'neonatolog'),
    ('Immunolog', 'immunolog'),
    ('Parazitolog', 'parazitolog')
ON CONFLICT (slug) DO NOTHING;

