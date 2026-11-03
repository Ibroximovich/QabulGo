import type { MetadataRoute } from 'next';

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: '*',
      allow: [
        '/',
        '/clinics',
        '/clinics/*',
        '/doctors',
        '/doctors/*',
      ],
      disallow: [
        '/admin',
        '/admin/*',
        '/clinic',
        '/clinic/*',
        '/dashboard',
        '/dashboard/*',
        '/profile',
        '/profile/*',
        '/appointments',
        '/appointments/*',
        '/login',
        '/register',
        '/verify-email',
        '/forgot-password',
        '/reset-password',
        '/api/*',
      ],
    },
    sitemap: 'https://qabulgo.uz/sitemap.xml',
    host: 'https://qabulgo.uz',
  };
}
