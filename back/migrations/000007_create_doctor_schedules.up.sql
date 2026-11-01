-- 000007_create_doctor_schedules.up.sql

-- 1. Doctor Weekly Slots table (Haftalik takrorlanuvchi qabul vaqtlari)
CREATE TABLE IF NOT EXISTS doctor_weekly_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_doctor_id UUID NOT NULL REFERENCES clinic_doctors(id) ON DELETE CASCADE,
    weekday INT NOT NULL,
    start_time VARCHAR(5) NOT NULL,
    end_time VARCHAR(5) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_weekly_slot_weekday CHECK (weekday BETWEEN 1 AND 7),
    CONSTRAINT chk_weekly_slot_time_order CHECK (start_time < end_time)
);

CREATE INDEX IF NOT EXISTS idx_doctor_weekly_slots_cd_weekday ON doctor_weekly_slots(clinic_doctor_id, weekday);
CREATE INDEX IF NOT EXISTS idx_doctor_weekly_slots_status ON doctor_weekly_slots(clinic_doctor_id, status);

-- 2. Doctor Date Slot Overrides table (Aniq sanadagi istisnolar va maxsus vaqtlar)
CREATE TABLE IF NOT EXISTS doctor_date_slot_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_doctor_id UUID NOT NULL REFERENCES clinic_doctors(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    start_time VARCHAR(5),
    end_time VARCHAR(5),
    status VARCHAR(20) NOT NULL DEFAULT 'closed',
    reason VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_override_status CHECK (status IN ('closed', 'available', 'day_off')),
    CONSTRAINT chk_override_time_order CHECK (
        status = 'day_off' OR (start_time IS NOT NULL AND end_time IS NOT NULL AND start_time < end_time)
    )
);

CREATE INDEX IF NOT EXISTS idx_doctor_date_overrides_cd_date ON doctor_date_slot_overrides(clinic_doctor_id, date);

