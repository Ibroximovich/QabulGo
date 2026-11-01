-- 000004_create_clinic_system_tables.up.sql

-- 1. Clinics table
CREATE TABLE IF NOT EXISTS clinics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    phone VARCHAR(50) NOT NULL,
    email VARCHAR(255),
    region VARCHAR(100) NOT NULL,
    district VARCHAR(100) NOT NULL,
    address VARCHAR(255) NOT NULL,
    logo_url VARCHAR(500),
    cover_image_url VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    rejection_reason TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_clinic_status CHECK (status IN ('draft', 'pending', 'approved', 'rejected', 'suspended'))
);

CREATE INDEX IF NOT EXISTS idx_clinics_status ON clinics(status);
CREATE INDEX IF NOT EXISTS idx_clinics_owner_id ON clinics(owner_id);
CREATE INDEX IF NOT EXISTS idx_clinics_region_district ON clinics(region, district);

-- 2. Clinic Members table (Klinika va uning ma'murlari/egasi)
CREATE TABLE IF NOT EXISTS clinic_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'owner',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_clinic_member UNIQUE(clinic_id, user_id),
    CONSTRAINT chk_clinic_member_role CHECK (role IN ('owner', 'admin'))
);

CREATE INDEX IF NOT EXISTS idx_clinic_members_user_id ON clinic_members(user_id);

-- 3. Specialties table (Shifokor mutaxassisliklari)
CREATE TABLE IF NOT EXISTS specialties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    icon_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_specialties_slug ON specialties(slug);

-- 4. Doctors table (Shifokorlarning umumiy bazasi)
CREATE TABLE IF NOT EXISTS doctors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    bio TEXT,
    experience_years INT NOT NULL DEFAULT 0,
    photo_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_doctors_name ON doctors(last_name, first_name);

-- 5. Clinic Doctors table (Klinika va shifokor mutaxassislik bog'lanmasi)
CREATE TABLE IF NOT EXISTS clinic_doctors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    specialty_id UUID NOT NULL REFERENCES specialties(id) ON DELETE RESTRICT,
    consultation_price NUMERIC(12, 2) NOT NULL DEFAULT 0,
    appointment_duration_minutes INT NOT NULL DEFAULT 30,
    is_active BOOLEAN NOT NULL DEFAULT true,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_clinic_doctor_specialty UNIQUE(clinic_id, doctor_id, specialty_id),
    CONSTRAINT chk_consultation_price CHECK (consultation_price >= 0),
    CONSTRAINT chk_appointment_duration CHECK (appointment_duration_minutes > 0)
);

CREATE INDEX IF NOT EXISTS idx_clinic_doctors_clinic ON clinic_doctors(clinic_id);
CREATE INDEX IF NOT EXISTS idx_clinic_doctors_specialty ON clinic_doctors(specialty_id);
CREATE INDEX IF NOT EXISTS idx_clinic_doctors_active ON clinic_doctors(clinic_id, is_active) WHERE deleted_at IS NULL;

-- 6. Clinic Working Hours table (Ish soatlari)
CREATE TABLE IF NOT EXISTS clinic_working_hours (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    weekday INT NOT NULL,
    opening_time VARCHAR(5) NOT NULL DEFAULT '08:00',
    closing_time VARCHAR(5) NOT NULL DEFAULT '18:00',
    is_closed BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT uq_clinic_weekday UNIQUE(clinic_id, weekday),
    CONSTRAINT chk_clinic_weekday CHECK (weekday BETWEEN 1 AND 7)
);

CREATE INDEX IF NOT EXISTS idx_clinic_working_hours_clinic ON clinic_working_hours(clinic_id);

-- 7. Clinic Verification Reviews table (Superadmin tekshiruv va audit tarixi)
CREATE TABLE IF NOT EXISTS clinic_verification_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    reviewed_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    decision VARCHAR(20) NOT NULL,
    reason TEXT,
    reviewed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_review_decision CHECK (decision IN ('approved', 'rejected', 'suspended'))
);

CREATE INDEX IF NOT EXISTS idx_clinic_reviews_clinic ON clinic_verification_reviews(clinic_id);

-- 8. Audit Logs table (Tizim audit amallari)
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);

-- 9. Specialties boshlang'ich seed ma'lumotlari
INSERT INTO specialties (name, slug) VALUES
    ('Terapevt', 'terapevt'),
    ('Nevrolog', 'nevrolog'),
    ('Kardiolog', 'kardiolog'),
    ('Urolog', 'urolog'),
    ('Pediatr', 'pediatr'),
    ('Dermatolog', 'dermatolog'),
    ('Ginekolog', 'ginekolog'),
    ('Oftalmolog', 'oftalmolog'),
    ('LOR (Otorinolaringolog)', 'lor'),
    ('Stomatolog', 'stomatolog'),
    ('Xirurg', 'xirurg'),
    ('Endokrinolog', 'endokrinolog'),
    ('Gastroenterolog', 'gastroenterolog'),
    ('Ortoped-Travmatolog', 'ortoped-travmatolog'),
    ('Allergolog', 'allergolog')
ON CONFLICT (slug) DO NOTHING;

