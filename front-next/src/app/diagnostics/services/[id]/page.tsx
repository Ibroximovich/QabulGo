import React, { Suspense } from 'react';
import type { Metadata } from 'next';
import { AppLayout } from '@/components/layout/AppLayout';
import { DiagnosticDetailPage } from '@/features/diagnostics/pages/DiagnosticDetailPage';

export const metadata: Metadata = {
  title: 'Diagnostika tekshiruviga yozilish — QabulGo',
  description: 'Tibbiy diagnostika uskunasi, xizmat narxi, tayyorgarlik qoidalari va qabul vaqtini tanlash.',
};

export default function Page({ params }: { params: { id: string } }) {
  return (
    <AppLayout>
      <Suspense
        fallback={
          <div className="max-w-5xl mx-auto p-6 space-y-6">
            <div className="h-48 bg-slate-100 rounded-3xl animate-pulse" />
            <div className="h-64 bg-slate-100 rounded-3xl animate-pulse" />
          </div>
        }
      >
        <DiagnosticDetailPage serviceId={params.id} />
      </Suspense>
    </AppLayout>
  );
}

