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
      };
    }

    const json = await res.json();
    const clinic = json?.data;

    if (!clinic) {
      return {
        title: 'Klinika topilmadi — QabulGo',
      };
    }

    const title = `${clinic.name} — QabulGo`;
    const desc =
      clinic.description ||
      `${clinic.name} — ${clinic.district ? `${clinic.district}, ` : ''}${clinic.region}. Shifokorlar va qabul vaqtlari.`;

    return {
      title,
      description: desc,
      openGraph: {
        title,
        description: desc,
        images: clinic.cover_image_url ? [clinic.cover_image_url] : undefined,
      },
    };
  } catch (e) {
    return {
      title: 'Klinika tafsilotlari — QabulGo',
    };
  }
}

export default function Page() {
  return <ClinicDetailPage />;
}

