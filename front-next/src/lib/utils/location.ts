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

/**
 * Xarita URL manzilidan (Yandex yoki Google Maps) koordinatalarni (latitude va longitude) ajratib olish
 */
export function extractCoordinatesFromUrl(url?: string | null): { latitude: number; longitude: number } | null {
  if (!url || typeof url !== 'string') return null;

  try {
    const decoded = decodeURIComponent(url);

    // 1. Yandex: ll=longitude,latitude yoki pt=longitude,latitude
    const llMatch = decoded.match(/[?&](?:ll|pt)=([0-9.]+)[,%]([0-9.]+)/i);
    if (llMatch) {
      const lon = parseFloat(llMatch[1]);
      const lat = parseFloat(llMatch[2]);
      if (!isNaN(lat) && !isNaN(lon) && lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180) {
        return { latitude: lat, longitude: lon };
      }
    }

    // 2. Google: @latitude,longitude,
    const atMatch = decoded.match(/@(-?[0-9.]+),(-?[0-9.]+)/);
    if (atMatch) {
      const lat = parseFloat(atMatch[1]);
      const lon = parseFloat(atMatch[2]);
      if (!isNaN(lat) && !isNaN(lon) && lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180) {
        return { latitude: lat, longitude: lon };
      }
    }

    // 3. Google Maps q=latitude,longitude
    const qMatch = decoded.match(/[?&]q=(-?[0-9.]+),(-?[0-9.]+)/i);
    if (qMatch) {
      const lat = parseFloat(qMatch[1]);
      const lon = parseFloat(qMatch[2]);
      if (!isNaN(lat) && !isNaN(lon) && lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180) {
        return { latitude: lat, longitude: lon };
      }
    }
  } catch {
    // ignore
  }

  return null;
}

/**
 * O'zbekiston viloyatlari markazlari koordinatalari (koordinatasi ko'rsatilmagan klinikalarga fallback sifatida)
 */
export const REGION_DEFAULT_COORDINATES: Record<string, { latitude: number; longitude: number }> = {
  'Toshkent shahri': { latitude: 41.311081, longitude: 69.240562 },
  'Toshkent viloyati': { latitude: 41.2858, longitude: 69.2035 },
  'Samarqand viloyati': { latitude: 39.6542, longitude: 66.9597 },
  'Farg‘ona viloyati': { latitude: 40.3842, longitude: 71.7843 },
  'Andijon viloyati': { latitude: 40.7821, longitude: 72.3442 },
  'Namangan viloyati': { latitude: 40.9983, longitude: 71.6726 },
  'Buxoro viloyati': { latitude: 39.7747, longitude: 64.4286 },
  'Qashqadaryo viloyati': { latitude: 38.8606, longitude: 65.7891 },
  'Surxondaryo viloyati': { latitude: 37.2242, longitude: 67.2783 },
  'Xorazm viloyati': { latitude: 41.5562, longitude: 60.6313 },
  'Navoiy viloyati': { latitude: 40.0844, longitude: 65.3792 },
  'Jizzax viloyati': { latitude: 40.1158, longitude: 67.8422 },
  'Sirdaryo viloyati': { latitude: 40.4983, longitude: 68.7842 },
  'Qoraqalpog‘iston Respublikasi': { latitude: 42.4619, longitude: 59.6166 },
};


