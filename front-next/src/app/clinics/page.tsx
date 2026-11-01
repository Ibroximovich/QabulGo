import React, { Suspense } from 'react';
import type { Metadata } from 'next';
import { ClinicSearchPage } from '@/features/search/pages/ClinicSearchPage';
import { ClinicCardSkeleton } from '@/features/search/components/SearchSkeletons';

export const metadata: Metadata = {
  title: 'Klinikalar — QabulGo',
  description:
    'O‘zbekistondagi xususiy klinikalar, tibbiy markazlar va diagnostika markazlari ro‘yxati. Eng yaqin klinikalarni toping va onlayn qabulga yoziling.',
  openGraph: {
    title: 'Klinikalar — QabulGo',
    description:
      'O‘zbekistondagi eng yaxshi xususiy klinikalarni qidirish, mutaxassisliklar va masofa bo‘yicha topish.',
    type: 'website',
  },
};

export default function Page() {
  return (
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
  );
}
