/**
 * Google Maps yoki Yandex Maps lokatsiya linkini qat'iy tekshiruvchi funksiya.
 * Faqat haqiqiy xarita havolalari (maps.google.com, goo.gl/maps, maps.app.goo.gl, yandex.uz/maps, yandex.ru/maps va h.k.) qabul qilinadi.
 */
export function isValidMapLocationUrl(url: string): boolean {
  if (!url || typeof url !== 'string') return false;
  const trimmed = url.trim();

  // URL formati bo'lishi kerak
  try {
    const parsed = new URL(trimmed.startsWith('http') ? trimmed : `https://${trimmed}`);
    const host = parsed.hostname.toLowerCase();

    // 1. Google Maps
    const isGoogleMaps =
      host.includes('google.') && (parsed.pathname.includes('/maps') || host.startsWith('maps.google.'));
    const isGoogleShort = host === 'goo.gl' && parsed.pathname.includes('/maps');
    const isGoogleApp = host === 'maps.app.goo.gl';

    // 2. Yandex Maps
    const isYandexMaps =
      host.includes('yandex.') && (parsed.pathname.includes('/maps') || parsed.pathname.includes('/navi'));
    const isYandexSub = host.startsWith('maps.yandex.');

    return isGoogleMaps || isGoogleShort || isGoogleApp || isYandexMaps || isYandexSub;
  } catch {
    return false;
  }
}

