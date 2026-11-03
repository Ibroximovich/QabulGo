
import type { MetadataRoute } from 'next';

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: '*',
      allow: '/',
      disallow: [
        '/admin/',
        '/dashboard/',
        '/profile/',
        '/appointments/',
        '/clinic/',
        '/login/',
        '/register/',
      ],
    },
    sitemap: 'https://qabulgo.uz/sitemap.xml',
  };
}
