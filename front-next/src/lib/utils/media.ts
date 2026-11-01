/**
 * Resolves media/image URL (e.g. /uploads/doctors/...) to full accessible URL.
 * If URL is already absolute (http://, https://, blob:, data:), returns it as is.
 */
export function getMediaUrl(url?: string | null): string {
  if (!url) return '';
  if (
    url.startsWith('http://') ||
    url.startsWith('https://') ||
    url.startsWith('blob:') ||
    url.startsWith('data:')
  ) {
    return url;
  }

  const rawBase =
    process.env.NEXT_PUBLIC_API_BASE_URL ||
    (typeof window !== 'undefined' ? (window as unknown as { __ENV_API_BASE_URL?: string }).__ENV_API_BASE_URL : '') ||
    'http://localhost:8080/api/v1';

  // Extract backend base origin, removing /api/v1
  const cleanBase = rawBase.replace(/\/+$/, '');
  const backendOrigin = cleanBase.replace(/\/api\/v1\/?$/, '');

  const normalizedPath = url.startsWith('/') ? url : `/${url}`;
  return `${backendOrigin}${normalizedPath}`;
}

