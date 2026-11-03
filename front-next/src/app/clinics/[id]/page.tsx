import type { Metadata } from 'next';
import { ClinicDetailPage } from '@/features/search/pages/ClinicDetailPage';

interface PageProps {
  params: { id: string };
}

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1';

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { id } = params;

  try {
    const res = await fetch(`${API_BASE}/public/clinics/${id}`, {
      next: { revalidate: 60 },
    });

    if (!res.ok) {
      return {
        title: 'Klinika topilmadi — QabulGo',
        robots: { index: false, follow: false },
      };
    }

    const json = await res.json();
    const clinic = json?.data;

    if (!clinic) {
      return {
        title: 'Klinika topilmadi — QabulGo',
        robots: { index: false, follow: false },
      };
    }

    const locationStr = clinic.district ? `${clinic.district}, ${clinic.region}` : clinic.region;
    const title = `${clinic.name} — ${locationStr} klinikasi | QabulGo`;
    const desc =
      clinic.description ||
      `${clinic.name} (${locationStr}) xususiy klinikasi. Shifokorlar ro‘yxati, qabul narxlari, manzili va onlayn navbatga yozilish.`;

    return {
      title,
      description: desc,
      alternates: {
        canonical: `/clinics/${id}`,
      },
      openGraph: {
        title,
        description: desc,
        url: `https://qabulgo.uz/clinics/${id}`,
        type: 'website',
        images: clinic.cover_image_url
          ? [clinic.cover_image_url]
          : clinic.logo_url
          ? [clinic.logo_url]
          : ['/logo-mark.png'],
      },
      twitter: {
        card: 'summary',
        title,
        description: desc,
        images: clinic.cover_image_url
          ? [clinic.cover_image_url]
          : clinic.logo_url
          ? [clinic.logo_url]
          : ['/logo-mark.png'],
      },
    };
  } catch (e) {
    return {
      title: 'Klinika tafsilotlari — QabulGo',
    };
  }
}

export default async function Page({ params }: PageProps) {
  const { id } = params;
  let clinic = null;
  let doctors = [];

  try {
    const [cRes, dRes] = await Promise.all([
      fetch(`${API_BASE}/public/clinics/${id}`, { next: { revalidate: 60 } }),
      fetch(`${API_BASE}/public/clinics/${id}/doctors`, { next: { revalidate: 60 } }),
    ]);

    if (cRes.ok) {
      const cJson = await cRes.json();
      clinic = cJson?.data || null;
    }
    if (dRes.ok) {
      const dJson = await dRes.json();
      doctors = dJson?.data || [];
    }
  } catch (e) {
    // Graceful fallback to client-side loading
  }

  const breadcrumbJsonLd = clinic
    ? {
        '@context': 'https://schema.org',
        '@type': 'BreadcrumbList',
        itemListElement: [
          {
            '@type': 'ListItem',
            position: 1,
            name: 'Bosh sahifa',
            item: 'https://qabulgo.uz',
          },
          {
            '@type': 'ListItem',
            position: 2,
            name: 'Klinikalar',
            item: 'https://qabulgo.uz/clinics',
          },
          {
            '@type': 'ListItem',
            position: 3,
            name: clinic.name,
            item: `https://qabulgo.uz/clinics/${clinic.id}`,
          },
        ],
      }
    : null;

  const medicalClinicJsonLd = clinic
    ? {
        '@context': 'https://schema.org',
        '@type': 'MedicalClinic',
        name: clinic.name,
        description: clinic.description || undefined,
        url: `https://qabulgo.uz/clinics/${clinic.id}`,
        telephone: clinic.phone || undefined,
        address: {
          '@type': 'PostalAddress',
          addressLocality: clinic.district || clinic.region || 'Termiz',
          addressRegion: clinic.region || 'Surxondaryo viloyati',
          streetAddress: clinic.address || undefined,
          addressCountry: 'UZ',
        },
        ...(clinic.latitude && clinic.longitude
          ? {
              geo: {
                '@type': 'GeoCoordinates',
                latitude: clinic.latitude,
                longitude: clinic.longitude,
              },
            }
          : {}),
        image: clinic.cover_image_url || clinic.logo_url || 'https://qabulgo.uz/logo-mark.png',
      }
    : null;

  return (
    <>
      {breadcrumbJsonLd && (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }}
        />
      )}
      {medicalClinicJsonLd && (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(medicalClinicJsonLd) }}
        />
      )}
      <ClinicDetailPage initialClinic={clinic} initialDoctors={doctors} />
    </>
  );
}
