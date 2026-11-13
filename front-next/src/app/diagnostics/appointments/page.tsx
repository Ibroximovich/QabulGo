import React, { Suspense } from 'react';
import type { Metadata } from 'next';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { AppLayout } from '@/components/layout/AppLayout';
import { DiagnosticAppointmentsPage } from '@/features/diagnostics/pages/DiagnosticAppointmentsPage';

export const metadata: Metadata = {
  title: 'Mening diagnostika qabullarim — QabulGo',
  description: 'MRT, MSKT, UZI va boshqa diagnostika qabullari ro‘yxati va tayyorgarlik eslatmalari.',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['patient']}>
      <AppLayout>
        <Suspense
          fallback={
            <div className="max-w-4xl mx-auto p-6 space-y-4">
              <div className="h-8 w-48 bg-slate-200 rounded animate-pulse" />
              <div className="h-44 bg-slate-100 rounded-3xl animate-pulse" />
            </div>
          }
        >
          <DiagnosticAppointmentsPage />
        </Suspense>
      </AppLayout>
    </ProtectedRoute>
  );
}

