-- 000015_add_expanded_diagnostic_categories.down.sql
-- Xavfsiz qaytarish (safe rollback): yangi kategoriya klinika xizmatlari yoki uskunalarida
-- ishlatilayotgan bo'lsa, uni o'chirib yubormaslik kafolatlanadi.

DELETE FROM diagnostic_categories
WHERE slug IN (
    'mammografiya',
    'densitometriya',
    'fluorografiya',
    'eeg',
    'enmg',
    'dopplerografiya',
    'holter-monitoring',
    'spirometriya',
    'kolposkopiya',
    'audiometriya',
    'oftalmologik-diagnostika',
    'pet-kt',
    'polisomnografiya',
    'scintigrafiya'
)
AND id NOT IN (SELECT category_id FROM diagnostic_equipments WHERE category_id IS NOT NULL)
AND id NOT IN (SELECT category_id FROM diagnostic_services WHERE category_id IS NOT NULL);

