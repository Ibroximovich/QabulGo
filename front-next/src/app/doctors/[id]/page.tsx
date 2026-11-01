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
      };
    }

    const json = await res.json();
    const doc = json?.data;

    if (!doc) {
      return {
        title: 'Shifokor topilmadi — QabulGo',
      };
    }

    const title = `${doc.first_name} ${doc.last_name} — QabulGo`;
    const desc =
      doc.bio ||
      `${doc.first_name} ${doc.last_name} — ${doc.primary_specialty?.name || 'Shifokor'}. Ish tajribasi: ${doc.experience_years || 0} yil. Ishlaydigan klinikalari va qabul narxlari.`;

    return {
      title,
      description: desc,
      openGraph: {
        title,
        description: desc,
        images: doc.photo_url ? [doc.photo_url] : undefined,
      },
    };
  } catch (e) {
    return {
      title: 'Shifokor tafsilotlari — QabulGo',
    };
  }
}

export default function Page() {
  return <DoctorDetailPage />;
}

