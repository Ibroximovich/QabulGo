'use client';

import React from 'react';
import { Skeleton } from '@/components/ui/Skeleton';

export function ClinicCardSkeleton() {
  return (
    <div className="bg-white rounded-2xl border border-slate-200/80 overflow-hidden flex flex-col p-0">
      <Skeleton className="h-44 w-full rounded-none" />
      <div className="p-5 pt-8 space-y-3">
        <Skeleton className="h-6 w-3/4 rounded-lg" />
        <Skeleton className="h-4 w-1/2 rounded-md" />
        <div className="flex gap-2 pt-2">
          <Skeleton className="h-5 w-20 rounded-md" />
          <Skeleton className="h-5 w-24 rounded-md" />
        </div>
        <div className="pt-4 border-t border-slate-100 flex justify-between items-center">
          <Skeleton className="h-4 w-28 rounded-md" />
          <Skeleton className="h-6 w-16 rounded-md" />
        </div>
      </div>
    </div>
  );
}

export function DoctorCardSkeleton() {
  return (
    <div className="bg-white rounded-2xl border border-slate-200/80 p-5 flex flex-col sm:flex-row gap-5">
      <Skeleton className="w-24 h-24 rounded-2xl shrink-0" />
      <div className="flex-1 space-y-2.5">
        <div className="flex gap-2">
          <Skeleton className="h-5 w-28 rounded-lg" />
          <Skeleton className="h-5 w-20 rounded-lg" />
        </div>
        <Skeleton className="h-6 w-1/2 rounded-md" />
        <Skeleton className="h-4 w-3/4 rounded-md" />
        <Skeleton className="h-10 w-full rounded-xl mt-2" />
      </div>
      <div className="sm:w-36 flex sm:flex-col justify-between items-end gap-3 pt-3 sm:pt-0 border-t sm:border-t-0 sm:border-l border-slate-100 sm:pl-4">
        <Skeleton className="h-6 w-24 rounded-md" />
        <Skeleton className="h-9 w-full rounded-xl" />
      </div>
    </div>
  );
}

