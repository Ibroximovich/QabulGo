import type { Metadata } from 'next';
import { DoctorDetailPage } from '@/features/search/pages/DoctorDetailPage';

interface PageProps {
  params: { id: string };
}

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1';

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { id } = params;

  try {
    const res = await fetch(`${API_BASE}/public/doctors/${id}`, {
      next: { revalidate: 60 },
    });

    if (!res.ok) {
      return {
        title: 'Shifokor topilmadi — QabulGo',
        robots: { index: false, follow: false },
      };
    }

    const json = await res.json();
    const doc = json?.data;

    if (!doc) {
      return {
        title: 'Shifokor topilmadi — QabulGo',
        robots: { index: false, follow: false },
      };
    }

    const specialty = doc.primary_specialty?.name || 'Shifokor';
    const title = `${doc.first_name} ${doc.last_name} (${specialty}) — Shifokor qabuli | QabulGo`;
    const desc =
      doc.bio ||
      `${doc.first_name} ${doc.last_name} — ${specialty}. Ish tajribasi: ${doc.experience_years || 0} yil. Ishlaydigan klinikalari, qabul narxlari va onlayn navbatga yozilish.`;

    return {
      title,
      description: desc,
      alternates: {
        canonical: `/doctors/${id}`,
      },
      openGraph: {
        title,
        description: desc,
        url: `https://qabulgo.uz/doctors/${id}`,
        type: 'profile',
        images: doc.photo_url ? [doc.photo_url] : ['/logo-mark.png'],
      },
      twitter: {
        card: 'summary',
        title,
        description: desc,
        images: doc.photo_url ? [doc.photo_url] : ['/logo-mark.png'],
      },
    };
  } catch (e) {
    return {
      title: 'Shifokor tafsilotlari — QabulGo',
    };
  }
}

export default async function Page({ params }: PageProps) {
  const { id } = params;
  let doctor = null;

  try {
    const res = await fetch(`${API_BASE}/public/doctors/${id}`, {
      next: { revalidate: 60 },
    });
    if (res.ok) {
      const json = await res.json();
      doctor = json?.data || null;
    }
  } catch (e) {
    // Graceful fallback to client-side load
  }

  const breadcrumbJsonLd = doctor
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
            name: 'Shifokorlar',
            item: 'https://qabulgo.uz/doctors',
          },
          {
            '@type': 'ListItem',
            position: 3,
            name: `${doctor.first_name} ${doctor.last_name}`,
            item: `https://qabulgo.uz/doctors/${doctor.id}`,
          },
        ],
      }
    : null;

  const physicianJsonLd = doctor
    ? {
        '@context': 'https://schema.org',
        '@type': 'Physician',
        name: `${doctor.first_name} ${doctor.last_name}`,
        description: doctor.bio || undefined,
        url: `https://qabulgo.uz/doctors/${doctor.id}`,
        medicalSpecialty: doctor.primary_specialty?.name || undefined,
        image: doctor.photo_url || 'https://qabulgo.uz/logo-mark.png',
        worksFor: doctor.workplaces?.map((w: { clinic_name: string; clinic_address?: string }) => ({
          '@type': 'MedicalClinic',
          name: w.clinic_name,
          address: w.clinic_address || undefined,
        })),
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
      {physicianJsonLd && (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(physicianJsonLd) }}
        />
      )}
      <DoctorDetailPage initialDoctor={doctor} />
    </>
  );
}
