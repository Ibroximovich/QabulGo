-- 000008_create_appointments.up.sql

CREATE TABLE IF NOT EXISTS appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE RESTRICT,
    clinic_doctor_id UUID NOT NULL REFERENCES clinic_doctors(id) ON DELETE RESTRICT,
    appointment_date DATE NOT NULL,
    start_time VARCHAR(5) NOT NULL,
    end_time VARCHAR(5) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    consultation_price NUMERIC(12, 2) NOT NULL DEFAULT 0,
    idempotency_key VARCHAR(255),
    cancellation_reason TEXT,
    cancelled_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    no_show_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_appointment_status CHECK (status IN ('confirmed', 'cancelled', 'completed', 'no_show')),
    CONSTRAINT chk_appointment_time_order CHECK (start_time < end_time),
    CONSTRAINT chk_appointment_price CHECK (consultation_price >= 0)
);

-- Search and lookup indices
CREATE INDEX IF NOT EXISTS idx_appointments_patient ON appointments(patient_id, appointment_date);
CREATE INDEX IF NOT EXISTS idx_appointments_clinic ON appointments(clinic_id, appointment_date);
CREATE INDEX IF NOT EXISTS idx_appointments_doctor ON appointments(doctor_id, appointment_date);
CREATE INDEX IF NOT EXISTS idx_appointments_clinic_doctor ON appointments(clinic_doctor_id, appointment_date);
CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);

-- Database-level Hard Double-Booking Prevention:
-- When status is 'confirmed' or 'completed', a doctor cannot have another active appointment at the same date and start_time across ANY clinic
CREATE UNIQUE INDEX IF NOT EXISTS uq_doctor_active_slot 
ON appointments(doctor_id, appointment_date, start_time) 
WHERE status IN ('confirmed', 'completed');

-- A patient cannot have two active appointments at the exact same date and start_time
CREATE UNIQUE INDEX IF NOT EXISTS uq_patient_active_slot
ON appointments(patient_id, appointment_date, start_time)
WHERE status IN ('confirmed', 'completed');

-- Idempotency key uniqueness per patient
CREATE UNIQUE INDEX IF NOT EXISTS uq_appointments_patient_idempotency
ON appointments(patient_id, idempotency_key)
WHERE idempotency_key IS NOT NULL;

