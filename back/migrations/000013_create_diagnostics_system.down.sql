-- 000013_create_diagnostics_system.down.sql
-- Revert diagnostics module tables in reverse dependency order

DROP TABLE IF EXISTS diagnostic_appointments CASCADE;
DROP TABLE IF EXISTS diagnostic_date_slot_overrides CASCADE;
DROP TABLE IF EXISTS diagnostic_weekly_slots CASCADE;
DROP TABLE IF EXISTS diagnostic_services CASCADE;
DROP TABLE IF EXISTS diagnostic_equipments CASCADE;
DROP TABLE IF EXISTS diagnostic_categories CASCADE;

