'use client';

import React, { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { diagnosticApi } from '../api/diagnosticApi';
import { CategoryFilterTabs } from '../components/CategoryFilterTabs';
import { DiagnosticServiceCard } from '../components/DiagnosticServiceCard';
import { Search, Activity, Sparkles, Filter, AlertCircle } from 'lucide-react';
import { Input } from '@/components/ui/Input';

export function DiagnosticsSearchPage() {
  const [selectedCategorySlug, setSelectedCategorySlug] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState<string>('');

  // 1. Fetch categories
  const { data: categoriesData, isLoading: isLoadingCategories } = useQuery({
    queryKey: ['diagnosticCategories'],
    queryFn: diagnosticApi.getCategories,
  });

  const categories = categoriesData?.data?.categories || [];

  // Find category ID from slug
  const selectedCategoryID = useMemo(() => {
    if (!selectedCategorySlug) return undefined;
    const cat = categories.find((c) => c.slug === selectedCategorySlug);
    return cat?.id;
  }, [selectedCategorySlug, categories]);

  // 2. Fetch services
  const {
    data: servicesData,
    isLoading: isLoadingServices,
    isError,
  } = useQuery({
    queryKey: ['diagnosticServices', selectedCategoryID],
    queryFn: () => diagnosticApi.getServices({ category_id: selectedCategoryID }),
  });

  const allServices = servicesData?.data?.services || [];

  // Filter services by search text (service name or clinic name)
  const filteredServices = useMemo(() => {
    if (!searchQuery.trim()) return allServices;
    const q = searchQuery.toLowerCase().trim();
    return allServices.filter(
      (s) =>
        s.name.toLowerCase().includes(q) ||
        s.clinic?.name.toLowerCase().includes(q) ||
        s.equipment?.name.toLowerCase().includes(q)
    );
  }, [allServices, searchQuery]);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Hero Header */}
      <div className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-slate-900 via-blue-950 to-slate-900 text-white p-6 sm:p-10 shadow-xl">
        <div className="relative z-10 max-w-2xl space-y-3">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white/10 backdrop-blur-md text-xs font-semibold text-blue-200 border border-white/10">
            <Sparkles className="w-3.5 h-3.5 text-blue-400" />
            Aniq va Tezkor Tibbiy Tekshiruvlar
          </div>
          <h1 className="text-2xl sm:text-4xl font-extrabold tracking-tight">
            MRT, MSKT, UZI va Laboratoriya
          </h1>
          <p className="text-sm sm:text-base text-slate-300">
            Eng zamonaviy tibbiy asbob-uskunalar bilan jihozlangan klinikalarni taqqoslang, bo‘sh vaqtni tanlang va qabulga navbatsiz yoziling.
          </p>
        </div>

        {/* Decorative background elements */}
        <div className="absolute -right-12 -bottom-12 w-64 h-64 bg-brand-blue/20 rounded-full blur-3xl pointer-events-none" />
        <div className="absolute right-20 top-10 w-40 h-40 bg-indigo-500/10 rounded-full blur-2xl pointer-events-none" />
      </div>

      {/* Filters & Search Bar */}
      <div className="space-y-4">
        <div className="flex flex-col sm:flex-row gap-3 items-stretch sm:items-center justify-between">
          <div className="relative flex-1 max-w-md">
            <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Tekshiruv yoki apparat nomini qidiring (masalan: Bosh miya MRT, Siemens)..."
              className="w-full pl-10 pr-4 py-2.5 rounded-xl border border-slate-200 bg-white text-sm text-slate-900 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-brand-blue/20 focus:border-brand-blue transition-all"
            />
          </div>

          <div className="text-xs font-medium text-slate-500">
            Jami: <strong className="text-slate-800">{filteredServices.length}</strong> ta xizmat topildi
          </div>
        </div>

        {/* Category Tabs */}
        <CategoryFilterTabs
          categories={categories}
          selectedSlug={selectedCategorySlug}
          onSelect={setSelectedCategorySlug}
          isLoading={isLoadingCategories}
        />
      </div>

      {/* Services List / Grid */}
      {isLoadingServices ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="h-64 bg-slate-100 rounded-2xl animate-pulse" />
          ))}
        </div>
      ) : isError ? (
        <div className="p-8 rounded-2xl bg-rose-50 border border-rose-200 text-center space-y-3">
          <AlertCircle className="w-8 h-8 text-rose-500 mx-auto" />
          <h3 className="text-base font-bold text-rose-900">Ma’lumotlarni yuklashda xatolik yuz berdi</h3>
          <p className="text-xs text-rose-700">Iltimos, sahifani yangilab qaytadan urinib ko‘ring.</p>
        </div>
      ) : filteredServices.length === 0 ? (
        <div className="p-12 rounded-3xl bg-slate-50 border border-slate-200/80 text-center space-y-3">
          <Activity className="w-12 h-12 text-slate-300 mx-auto" />
          <h3 className="text-base font-bold text-slate-800">Ushbu parametrlar bo‘yicha diagnostika xizmatlari topilmadi</h3>
          <p className="text-xs text-slate-500 max-w-sm mx-auto">
            Boshqa kategoriya tanlab ko‘ring yoki qidiruv so‘zini o‘zgartiring.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredServices.map((svc) => (
            <DiagnosticServiceCard key={svc.id} service={svc} />
          ))}
        </div>
      )}
    </div>
  );
}

