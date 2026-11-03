import React, { Suspense } from 'react';
import type { Metadata } from 'next';
import { ClinicSearchPage } from '@/features/search/pages/ClinicSearchPage';
import { ClinicCardSkeleton } from '@/features/search/components/SearchSkeletons';

export const metadata: Metadata = {
  title: 'Termiz va O‘zbekiston xususiy klinikalari — QabulGo',
  description:
    'Termiz va O‘zbekiston bo‘yicha rasmiy litsenziyalangan barcha xususiy klinikalar, tibbiy markazlar va laboratoriyalar. Aniq manzillar, shifokorlar va onlayn navbatga yozilish.',
  alternates: {
    canonical: '/clinics',
  },
  openGraph: {
    title: 'Termiz va O‘zbekiston xususiy klinikalari — QabulGo',
    description:
      'Termiz va O‘zbekistondagi eng yaxshi xususiy klinikalarni qidirish, mutaxassisliklar va geolokatsiya bo‘yicha topish.',
    url: 'https://qabulgo.uz/clinics',
    type: 'website',
  },
  twitter: {
    card: 'summary',
    title: 'Termiz va O‘zbekiston xususiy klinikalari — QabulGo',
    description:
      'Termiz va O‘zbekistondagi eng yaxshi xususiy klinikalarni qidirish, mutaxassisliklar va geolokatsiya bo‘yicha topish.',
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
      name: 'Klinikalar',
      item: 'https://qabulgo.uz/clinics',
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
          <div className="max-w-7xl mx-auto p-6 sm:p-8 space-y-6">
            <div className="h-10 bg-slate-200 rounded-xl w-1/3 animate-pulse" />
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <ClinicCardSkeleton />
              <ClinicCardSkeleton />
              <ClinicCardSkeleton />
            </div>
          </div>
        }
      >
        <ClinicSearchPage />
      </Suspense>
    </>
  );
}
