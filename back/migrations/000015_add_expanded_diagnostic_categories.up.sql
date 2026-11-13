-- 000015_add_expanded_diagnostic_categories.up.sql
-- Diagnostika moduli: 14 ta yangi diagnostika kategoriyalarini qo'shish

INSERT INTO diagnostic_categories (name, slug, icon_url, description)
VALUES 
    ('Mammografiya', 'mammografiya', '', 'Ko‘krak bezi rentgenologik tekshiruvi'),
    ('Densitometriya', 'densitometriya', '', 'Suyak to‘qimasi zichligini aniqlash tekshiruvi'),
    ('Fluorografiya', 'fluorografiya', '', 'O‘pka a''zolarini profilaktik rentgen tekshiruvi'),
    ('EEG', 'eeg', '', 'Bosh miya bioelektrik faolligi elektroensefalografiyasi'),
    ('ENMG', 'enmg', '', 'Periferik asab va mushaklar elektromioneyrografiyasi'),
    ('Dopplerografiya', 'dopplerografiya', '', 'Qon tomirlari ultratovush dopplerografiyasi (UZDG)'),
    ('Holter monitoring', 'holter-monitoring', '', 'EKG va qon bosimini sutkalik monitoring qilish'),
    ('Spirometriya', 'spirometriya', '', 'Tashqi nafas funksiyasini tekshirish'),
    ('Kolposkopiya', 'kolposkopiya', '', 'Bachadon bo‘yni to‘qimalarini mikroskop ostida tekshirish'),
    ('Audiometriya', 'audiometriya', '', 'Eshitish qobiliyatini tekshirish va o‘lchash'),
    ('Oftalmologik diagnostika', 'oftalmologik-diagnostika', '', 'Ko‘z tubi, ko‘rish o‘tkirligi va ko‘z ichi bosimi diagnostikasi'),
    ('PET/KT', 'pet-kt', '', 'Pozitron-emissiya va kompyuter tomografiyasi tekshiruvi'),
    ('Polisomnografiya', 'polisomnografiya', '', 'Uyqu buzilishlari va uyqu apnoesi diagnostikasi'),
    ('Scintigrafiya', 'scintigrafiya', '', 'Radioizotop diagnostika tekshiruvlari')
ON CONFLICT (slug) DO NOTHING;

