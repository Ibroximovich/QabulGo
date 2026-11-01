import React, { Suspense } from 'react';
import type { Metadata } from 'next';
import { DoctorSearchPage } from '@/features/search/pages/DoctorSearchPage';
import { DoctorCardSkeleton } from '@/features/search/components/SearchSkeletons';

export const metadata: Metadata = {
  title: 'Shifokorlar — QabulGo',
  description:
    'O‘zbekistondagi malakali shifokorlar ro‘yxati. Mutaxassislik, ish staji, qabul narxi va klinikalar bo‘yicha eng yaxshi shifokorni toping.',
  openGraph: {
    title: 'Shifokorlar — QabulGo',
    description:
      'Malakali shifokorlar, qabul narxlari va klinika manzillari bo‘yicha qidiruv.',
    type: 'website',
  },
};

export default function Page() {
  return (
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
  );
}
