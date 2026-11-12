-- 000013_create_diagnostics_system.up.sql
-- Diagnostics module: Categories, Equipments, Services, Schedules, and Appointments

-- 1. Diagnostic Categories table (MRT, MSKT, UZI, Rentgen, EKG, Laboratoriya, etc.)
CREATE TABLE IF NOT EXISTS diagnostic_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    icon_url VARCHAR(500),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_diag_categories_slug ON diagnostic_categories(slug);

-- 2. Diagnostic Equipments / Rooms table (Klinikaning apparatlari yoki kabinetlari)
CREATE TABLE IF NOT EXISTS diagnostic_equipments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES diagnostic_categories(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    room_number VARCHAR(50),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_diag_equipments_clinic_id ON diagnostic_equipments(clinic_id);
CREATE INDEX IF NOT EXISTS idx_diag_equipments_category_id ON diagnostic_equipments(category_id);
CREATE INDEX IF NOT EXISTS idx_diag_equipments_is_active ON diagnostic_equipments(is_active);

-- 3. Diagnostic Services table (Aniq tekshiruv xizmati, narxi, davomiyligi va tayyorgarlik qoidalari)
CREATE TABLE IF NOT EXISTS diagnostic_services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES diagnostic_categories(id) ON DELETE RESTRICT,
    equipment_id UUID NOT NULL REFERENCES diagnostic_equipments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    price NUMERIC(12, 2) NOT NULL DEFAULT 0,
    duration_minutes INT NOT NULL DEFAULT 30,
    preparation_rules TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_diag_service_price CHECK (price >= 0),
    CONSTRAINT chk_diag_service_duration CHECK (duration_minutes > 0)
);

CREATE INDEX IF NOT EXISTS idx_diag_services_clinic_id ON diagnostic_services(clinic_id);
CREATE INDEX IF NOT EXISTS idx_diag_services_category_id ON diagnostic_services(category_id);
CREATE INDEX IF NOT EXISTS idx_diag_services_equipment_id ON diagnostic_services(equipment_id);
CREATE INDEX IF NOT EXISTS idx_diag_services_is_active ON diagnostic_services(is_active);

-- 4. Diagnostic Weekly Slots table (Uskunaning haftalik takrorlanuvchi ish vaqtlari)
CREATE TABLE IF NOT EXISTS diagnostic_weekly_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_id UUID NOT NULL REFERENCES diagnostic_equipments(id) ON DELETE CASCADE,
    weekday INT NOT NULL,
    start_time VARCHAR(5) NOT NULL,
    end_time VARCHAR(5) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_diag_weekly_slot_weekday CHECK (weekday BETWEEN 1 AND 7),
    CONSTRAINT chk_diag_weekly_slot_time_order CHECK (start_time < end_time),
    CONSTRAINT chk_diag_weekly_slot_status CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX IF NOT EXISTS idx_diag_weekly_slots_eq_weekday ON diagnostic_weekly_slots(equipment_id, weekday);
CREATE INDEX IF NOT EXISTS idx_diag_weekly_slots_status ON diagnostic_weekly_slots(equipment_id, status);

-- 5. Diagnostic Date Slot Overrides table (Aniq sanadagi texnik profilaktika yoki maxsus ish vaqtlari)
CREATE TABLE IF NOT EXISTS diagnostic_date_slot_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_id UUID NOT NULL REFERENCES diagnostic_equipments(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    start_time VARCHAR(5),
    end_time VARCHAR(5),
    status VARCHAR(20) NOT NULL DEFAULT 'closed',
    reason VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_diag_override_status CHECK (status IN ('closed', 'available', 'day_off')),
    CONSTRAINT chk_diag_override_time_order CHECK (
        status = 'day_off' OR (start_time IS NOT NULL AND end_time IS NOT NULL AND start_time < end_time)
    )
);

CREATE INDEX IF NOT EXISTS idx_diag_date_overrides_eq_date ON diagnostic_date_slot_overrides(equipment_id, date);

-- 6. Diagnostic Appointments table (Diagnostika qabullari)
CREATE TABLE IF NOT EXISTS diagnostic_appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    service_id UUID NOT NULL REFERENCES diagnostic_services(id) ON DELETE RESTRICT,
    equipment_id UUID NOT NULL REFERENCES diagnostic_equipments(id) ON DELETE RESTRICT,
    appointment_date DATE NOT NULL,
    start_time VARCHAR(5) NOT NULL,
    end_time VARCHAR(5) NOT NULL,
    price NUMERIC(12, 2) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    preparation_acknowledged BOOLEAN NOT NULL DEFAULT false,
    idempotency_key VARCHAR(255),
    cancellation_reason TEXT,
    cancelled_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    no_show_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_diag_appointment_status CHECK (status IN ('confirmed', 'cancelled', 'completed', 'no_show')),
    CONSTRAINT chk_diag_appointment_time_order CHECK (start_time < end_time),
    CONSTRAINT chk_diag_appointment_price CHECK (price >= 0)
);

-- Search and lookup indices
CREATE INDEX IF NOT EXISTS idx_diag_appts_patient ON diagnostic_appointments(patient_id, appointment_date);
CREATE INDEX IF NOT EXISTS idx_diag_appts_clinic ON diagnostic_appointments(clinic_id, appointment_date);
CREATE INDEX IF NOT EXISTS idx_diag_appts_service ON diagnostic_appointments(service_id, appointment_date);
CREATE INDEX IF NOT EXISTS idx_diag_appts_equipment ON diagnostic_appointments(equipment_id, appointment_date);
CREATE INDEX IF NOT EXISTS idx_diag_appts_status ON diagnostic_appointments(status);

-- Database-level Hard Double-Booking Prevention:
-- When status is 'confirmed' or 'completed', an equipment cannot have another booking at the same date and start_time
CREATE UNIQUE INDEX IF NOT EXISTS uq_equipment_active_slot 
ON diagnostic_appointments(equipment_id, appointment_date, start_time) 
WHERE status IN ('confirmed', 'completed');

-- A patient cannot have two active diagnostic appointments at the exact same date and start_time
CREATE UNIQUE INDEX IF NOT EXISTS uq_diag_patient_active_slot
ON diagnostic_appointments(patient_id, appointment_date, start_time)
WHERE status IN ('confirmed', 'completed');

-- Idempotency key uniqueness per patient
CREATE UNIQUE INDEX IF NOT EXISTS uq_diag_appointments_patient_idempotency
ON diagnostic_appointments(patient_id, idempotency_key)
WHERE idempotency_key IS NOT NULL;

-- Initial Seed: Standard Diagnostic Categories
INSERT INTO diagnostic_categories (name, slug, icon_url, description)
VALUES 
    ('MRT', 'mrt', '', 'Magnit-rezonans tomografiya tekshiruvlari'),
    ('MSKT', 'mskt', '', 'Multispiral kompyuter tomografiya tekshiruvlari'),
    ('UZI', 'uzi', '', 'Ultratovush tekshiruvlari'),
    ('Rentgen', 'rentgen', '', 'Rentgenografiya va rentgenoskopiya tekshiruvlari'),
    ('EKG va EXO-KG', 'ekg-exo-kg', '', 'Yurak faoliyati va exokardiografiya diagnostikasi'),
    ('Laboratoriya', 'laboratoriya', '', 'Klinik, biokimyoviy va immunoferment laborator tahlillar'),
    ('Endoskopiya', 'endoskopiya', '', 'Gastroskopiya, kolonoskopiya va boshqa endoskopik tekshiruvlar')
ON CONFLICT (slug) DO NOTHING;

