-- 000004_create_clinic_system_tables.down.sql

DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS clinic_verification_reviews CASCADE;
DROP TABLE IF EXISTS clinic_working_hours CASCADE;
DROP TABLE IF EXISTS clinic_doctors CASCADE;
DROP TABLE IF EXISTS doctors CASCADE;
DROP TABLE IF EXISTS specialties CASCADE;
DROP TABLE IF EXISTS clinic_members CASCADE;
DROP TABLE IF EXISTS clinics CASCADE;

