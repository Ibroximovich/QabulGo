import React, { Suspense } from 'react';
import type { Metadata } from 'next';
import { AppLayout } from '@/components/layout/AppLayout';
import { DiagnosticsSearchPage } from '@/features/diagnostics/pages/DiagnosticsSearchPage';

export const metadata: Metadata = {
  title: 'MRT, MSKT, UZI va Laboratoriya diagnostikasi — QabulGo',
  description:
    'O‘zbekistondagi eng zamonaviy apparatlar bilan jihozlangan xususiy klinikalarda MRT, MSKT, UZI, Rentgen va laboratoriya tekshiruvlariga onlayn navbatsiz yoziling.',
};

export default function Page() {
  return (
    <AppLayout>
      <Suspense
        fallback={
          <div className="max-w-7xl mx-auto p-6 space-y-6">
            <div className="h-44 bg-slate-100 rounded-3xl animate-pulse" />
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {[...Array(6)].map((_, i) => (
                <div key={i} className="h-64 bg-slate-100 rounded-2xl animate-pulse" />
              ))}
            </div>
          </div>
        }
      >
        <DiagnosticsSearchPage />
      </Suspense>
    </AppLayout>
  );
}

