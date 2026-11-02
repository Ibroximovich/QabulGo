import React from 'react';

export default function RootLoading() {
  return (
    <div className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-white/80 backdrop-blur-xs transition-opacity duration-200">
      {/* Top indeterminate progress bar */}
      <div className="fixed top-0 left-0 right-0 h-1 bg-brand-blue/20 overflow-hidden z-50">
        <div className="h-full bg-brand-blue animate-[shimmer_1.5s_infinite] w-full" />
      </div>

      <div className="flex flex-col items-center gap-3">
        <div className="relative w-12 h-12">
          <div className="w-12 h-12 rounded-full border-3 border-brand-blue/20 border-t-brand-blue animate-spin" />
        </div>
        <p className="text-xs font-semibold text-slate-500 animate-pulse tracking-wide">
          Yuklanmoqda...
        </p>
      </div>
    </div>
  );
}
