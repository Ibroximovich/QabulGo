import React, { Suspense } from 'react';
import type { Metadata } from 'next';
import { DoctorSearchPage } from '@/features/search/pages/DoctorSearchPage';
import { DoctorCardSkeleton } from '@/features/search/components/SearchSkeletons';

export const metadata: Metadata = {
  title: 'Termiz va O‘zbekiston shifokorlari — QabulGo',
  description:
    'Termiz va O‘zbekiston bo‘ylab malakali shifokorlar (nevropatolog, urolog, kardiolog, pediatr, stomatolog va b.). Ish staji, klinikalar, qabul narxlari va onlayn navbatga yozilish.',
  alternates: {
    canonical: '/doctors',
  },
  openGraph: {
    title: 'Termiz va O‘zbekiston shifokorlari — QabulGo',
    description:
      'Malakali shifokorlar, mutaxassisliklar, qabul narxlari va klinikalar bo‘yicha qidiruv.',
    url: 'https://qabulgo.uz/doctors',
    type: 'website',
  },
  twitter: {
    card: 'summary',
    title: 'Termiz va O‘zbekiston shifokorlari — QabulGo',
    description:
      'Malakali shifokorlar, mutaxassisliklar, qabul narxlari va klinikalar bo‘yicha qidiruv.',
  },
};

const breadcrumbJsonLd = {
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
  ],
};

export default function Page() {
  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }}
      />
      <Suspense
        fallback={
          <div className="max-w-5xl mx-auto p-6 sm:p-8 space-y-6">
            <div className="h-10 bg-slate-200 rounded-xl w-1/3 animate-pulse" />
            <div className="space-y-4">
              <DoctorCardSkeleton />
              <DoctorCardSkeleton />
              <DoctorCardSkeleton />
            </div>
          </div>
        }
      >
        <DoctorSearchPage />
      </Suspense>
    </>
  );
}
