'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { DoctorFilterBar, DoctorFiltersState } from '../components/DoctorFilterBar';
import { DoctorCard } from '../components/DoctorCard';
import { DoctorCardSkeleton } from '../components/SearchSkeletons';
import { EmptySearchState } from '../components/EmptySearchState';
import { searchApi } from '../api/searchApi';
import { PublicDoctorItem, RegionItem, SpecialtyItem } from '../types';
import { ChevronLeft, ChevronRight, Stethoscope } from 'lucide-react';
import { Button } from '@/components/ui/Button';

export function DoctorSearchPage() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  // URL Query State
  const initialSearch = searchParams.get('search') || '';
  const initialRegion = searchParams.get('region') || '';
  const initialDistrict = searchParams.get('district') || '';
  const initialSpecialty = searchParams.get('specialty_id') || '';
  const initialMinPrice = searchParams.get('min_price') || '';
  const initialMaxPrice = searchParams.get('max_price') || '';
  const initialSort = (searchParams.get('sort') as any) || 'name_asc';
  const initialPage = parseInt(searchParams.get('page') || '1', 10);

  const [filters, setFilters] = useState<DoctorFiltersState>({
    search: initialSearch,
    region: initialRegion,
    
    district: initialDistrict,
    specialtyId: initialSpecialty,
    minPrice: initialMinPrice,
    maxPrice: initialMaxPrice,
    sort: initialSort,
  });

  const [page, setPage] = useState<number>(initialPage);
  const [limit] = useState<number>(12);
  const [total, setTotal] = useState<number>(0);

  // Catalogs
  const [regions, setRegions] = useState<RegionItem[]>([]);
  const [specialties, setSpecialties] = useState<SpecialtyItem[]>([]);

  // Results & UI State
  const [doctors, setDoctors] = useState<PublicDoctorItem[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // Load catalogs on mount
  useEffect(() => {
    let active = true;
    async function loadCatalogs() {
      try {
        const [regRes, specRes] = await Promise.all([
          searchApi.getRegionsCatalog(),
          searchApi.getSpecialties(),
        ]);
        if (active) {
          if (regRes.data) setRegions(regRes.data);
          if (specRes.data) setSpecialties(specRes.data);
        }
      } catch (err) {
        console.error('Katalogni yuklashda xatolik:', err);
      }
    }
    loadCatalogs();
    return () => {
      active = false;
    };
  }, []);

  // Sync URL query params with state
  const updateUrlParams = useCallback(
    (newFilters: DoctorFiltersState, newPage: number) => {
      const params = new URLSearchParams();
      if (newFilters.search) params.set('search', newFilters.search);
      if (newFilters.region) params.set('region', newFilters.region);
      if (newFilters.district) params.set('district', newFilters.district);
      if (newFilters.specialtyId) params.set('specialty_id', newFilters.specialtyId);
      if (newFilters.minPrice) params.set('min_price', newFilters.minPrice);
      if (newFilters.maxPrice) params.set('max_price', newFilters.maxPrice);
      if (newFilters.sort && newFilters.sort !== 'name_asc') params.set('sort', newFilters.sort);
      if (newPage > 1) params.set('page', String(newPage));

      const queryString = params.toString();
      const targetUrl = queryString ? `${pathname}?${queryString}` : pathname;
      router.replace(targetUrl, { scroll: false });
    },
    [pathname, router]
  );

  // Fetch doctors on filter/page change
  useEffect(() => {
    let cancelled = false;

    async function fetchDoctors() {
      setIsLoading(true);
      setError(null);

      try {
        const minP = filters.minPrice ? parseFloat(filters.minPrice) : undefined;
        const maxP = filters.maxPrice ? parseFloat(filters.maxPrice) : undefined;

        const res = await searchApi.searchDoctors({
          search: filters.search.trim() || undefined,
          region: filters.region || undefined,
          district: filters.district || undefined,
          specialty_id: filters.specialtyId || undefined,
          min_price: !isNaN(minP as any) ? minP : undefined,
          max_price: !isNaN(maxP as any) ? maxP : undefined,
          sort: filters.sort,
          page,
          limit,
        });

        if (!cancelled) {
          setDoctors(res.data || []);
          setTotal(res.meta?.total || 0);
        }
      } catch (err: any) {
        if (!cancelled) {
          console.error('Shifokorlarni qidirishda xatolik:', err);
          setError('Shifokorlarni yuklashda xatolik yuz berdi. Iltimos qayta urinib ko‘ring.');
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    }

    const timer = setTimeout(() => {
      fetchDoctors();
    }, 250);

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [filters, page, limit]);

  // Handle filter changes
  const handleFilterChange = (partial: Partial<DoctorFiltersState>) => {
    const updated = { ...filters, ...partial };
    setFilters(updated);
    setPage(1);
    updateUrlParams(updated, 1);
  };

  // Reset filters
  const handleReset = () => {
    const resetState: DoctorFiltersState = {
      search: '',
      region: '',
      district: '',
      specialtyId: '',
      minPrice: '',
      maxPrice: '',
      sort: 'name_asc',
    };
    setFilters(resetState);
    setPage(1);
    updateUrlParams(resetState, 1);
  };

  const totalPages = Math.max(1, Math.ceil(total / limit));

  return (
    <AppLayout>
      <div className="space-y-6">
        {/* Page Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 text-brand-blue text-xs font-bold uppercase tracking-wider mb-1">
              <Stethoscope className="w-4 h-4" />
              <span>Shifokorlar katalogi</span>
            </div>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-brand-navy tracking-tight">
              Malakali shifokorlarni topish
            </h1>
            <p className="text-sm text-slate-500 mt-1 max-w-2xl">
              Mutaxassislik, ish tajribasi, qabul narxi va klinikalar bo‘yicha o‘zingizga mos eng yaxshi shifokorni tanlang.
            </p>
          </div>

          {/* Results Counter Badge */}
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center px-3.5 py-1.5 rounded-full bg-white border border-slate-200/80 text-xs font-semibold text-brand-navy shadow-xs">
              Jami: <strong className="text-brand-blue ml-1 font-bold">{total}</strong> ta natija
            </span>
          </div>
        </div>

        {/* Filters Bar */}
        <DoctorFilterBar
          filters={filters}
          onFilterChange={handleFilterChange}
          onReset={handleReset}
          regions={regions}
          specialties={specialties}
          isLoading={isLoading}
        />

        {/* Content Display */}
        {error ? (
          <div className="p-6 bg-red-50 text-red-700 rounded-2xl border border-red-200 text-center">
            <p className="font-semibold">{error}</p>
            <Button variant="outline" size="sm" onClick={() => handleFilterChange({})} className="mt-3">
              Qayta urinish
            </Button>
          </div>
        ) : isLoading ? (
          <div className="space-y-4">
            {Array.from({ length: 6 }).map((_, i) => (
              <DoctorCardSkeleton key={i} />
            ))}
          </div>
        ) : doctors.length === 0 ? (
          <EmptySearchState onReset={handleReset} />
        ) : (
          <div className="space-y-4">
            <div className="grid grid-cols-1 gap-4">
              {doctors.map((doctor) => (
                <DoctorCard key={doctor.clinic_doctor_id} doctor={doctor} />
              ))}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="pt-6 flex items-center justify-between border-t border-slate-200 max-w-md mx-auto">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => {
                    const p = Math.max(1, page - 1);
                    setPage(p);
                    updateUrlParams(filters, p);
                  }}
                  className="gap-1.5"
                >
                  <ChevronLeft className="w-4 h-4" />
                  <span>Oldingi</span>
                </Button>

                <span className="text-xs font-semibold text-slate-500">
                  {page} / {totalPages} sahifa
                </span>

                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= totalPages}
                  onClick={() => {
                    const p = Math.min(totalPages, page + 1);
                    setPage(p);
                    updateUrlParams(filters, p);
                  }}
                  className="gap-1.5"
                >
                  <span>Keyingi</span>
                  <ChevronRight className="w-4 h-4" />
                </Button>
              </div>
            )}
          </div>
        )}
      </div>
    </AppLayout>
  );
}

