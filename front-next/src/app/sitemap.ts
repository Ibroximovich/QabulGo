import type { MetadataRoute } from 'next';

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1';
const baseUrl = 'https://qabulgo.uz';

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const routes: MetadataRoute.Sitemap = [
    {
      url: baseUrl,
      lastModified: new Date(),
      changeFrequency: 'daily',
      priority: 1.0,
    },
    {
      url: `${baseUrl}/clinics`,
      lastModified: new Date(),
      changeFrequency: 'daily',
      priority: 0.9,
    },
    {
      url: `${baseUrl}/doctors`,
      lastModified: new Date(),
      changeFrequency: 'daily',
      priority: 0.9,
    },
  ];

  try {
    const clinicsRes = await fetch(`${API_BASE}/public/clinics?limit=100`, {
      next: { revalidate: 3600 },
    });
    if (clinicsRes.ok) {
      const json = await clinicsRes.json();
      const clinics = json?.data || [];
      clinics.forEach((c: { id: string; updated_at?: string; created_at?: string }) => {
        if (c?.id) {
          routes.push({
            url: `${baseUrl}/clinics/${c.id}`,
            lastModified: c.updated_at ? new Date(c.updated_at) : new Date(),
            changeFrequency: 'weekly',
            priority: 0.8,
          });
        }
      });
    }
  } catch (e) {
    // Silent fallback to core routes
  }

  try {
    const doctorsRes = await fetch(`${API_BASE}/public/doctors?limit=100`, {
      next: { revalidate: 3600 },
    });
    if (doctorsRes.ok) {
      const json = await doctorsRes.json();
      const doctors = json?.data || [];
      doctors.forEach((d: { id: string; updated_at?: string; created_at?: string }) => {
        if (d?.id) {
          routes.push({
            url: `${baseUrl}/doctors/${d.id}`,
            lastModified: d.updated_at ? new Date(d.updated_at) : new Date(),
            changeFrequency: 'weekly',
            priority: 0.8,
          });
        }
      });
    }
  } catch (e) {
    // Silent fallback to core routes
  }

  return routes;
}
