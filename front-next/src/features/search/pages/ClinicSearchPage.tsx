'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { AppLayout } from '@/components/layout/AppLayout';
import { ClinicFilterBar, ClinicFiltersState } from '../components/ClinicFilterBar';
import { ClinicCard } from '../components/ClinicCard';
import { InteractiveMap } from '@/components/map/InteractiveMap';
import { ClinicCardSkeleton } from '../components/SearchSkeletons';
import { EmptySearchState } from '../components/EmptySearchState';
import { ClinicMapDetailCard } from '../components/ClinicMapDetailCard';
import { searchApi } from '../api/searchApi';
import { PublicClinicItem, NearbyClinicItem, RegionItem, SpecialtyItem } from '../types';
import { ChevronLeft, ChevronRight, Building2, MapPin } from 'lucide-react';
import { Button } from '@/components/ui/Button';

export function ClinicSearchPage() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  // URL Query State
  const initialSearch = searchParams.get('search') || '';
  const initialRegion = searchParams.get('region') || '';
  const initialDistrict = searchParams.get('district') || '';
  const initialSpecialty = searchParams.get('specialty_id') || '';
  const initialPage = parseInt(searchParams.get('page') || '1', 10);

  const [filters, setFilters] = useState<ClinicFiltersState>({
    search: initialSearch,
    region: initialRegion,
    district: initialDistrict,
    specialtyId: initialSpecialty,
  });

  const [page, setPage] = useState<number>(initialPage);
  const [limit] = useState<number>(50);
  const [total, setTotal] = useState<number>(0);

  // Nearby Mode State
  const [isNearbyMode, setIsNearbyMode] = useState<boolean>(false);
  const [userLocation, setUserLocation] = useState<{ latitude: number; longitude: number } | null>(null);
  const [nearbyRadiusKm, setNearbyRadiusKm] = useState<number>(10);

  // Cached Catalogs via React Query (30 min staleTime)
  const { data: regions = [] } = useQuery({
    queryKey: ['catalogRegions'],
    queryFn: async () => {
      const res = await searchApi.getRegionsCatalog();
      return res.data || [];
    },
    staleTime: 30 * 60 * 1000,
  });

  const { data: specialties = [] } = useQuery({
    queryKey: ['catalogSpecialties'],
    queryFn: async () => {
      const res = await searchApi.getSpecialties();
      return res.data || [];
    },
    staleTime: 30 * 60 * 1000,
  });

  // Results & UI State
  const [clinics, setClinics] = useState<(PublicClinicItem | NearbyClinicItem)[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedClinicId, setSelectedClinicId] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'list' | 'map' | 'split'>('map');

  const selectedClinic = selectedClinicId
    ? clinics.find((c) => c.id === selectedClinicId) || null
    : null;

  // Sync URL query params with state
  const updateUrlParams = useCallback(
    (newFilters: ClinicFiltersState, newPage: number) => {
      const params = new URLSearchParams();
      if (newFilters.search) params.set('search', newFilters.search);
      if (newFilters.region) params.set('region', newFilters.region);
      if (newFilters.district) params.set('district', newFilters.district);
      if (newFilters.specialtyId) params.set('specialty_id', newFilters.specialtyId);
      if (newPage > 1) params.set('page', String(newPage));

      const queryString = params.toString();
      const targetUrl = queryString ? `${pathname}?${queryString}` : pathname;
      router.replace(targetUrl, { scroll: false });
    },
    [pathname, router]
  );

  // Debounced or direct search trigger
  useEffect(() => {
    let cancelled = false;

    async function fetchClinics() {
      setIsLoading(true);
      setError(null);

      try {
        if (isNearbyMode && userLocation) {
          const res = await searchApi.searchNearbyClinics({
            latitude: userLocation.latitude,
            longitude: userLocation.longitude,
            radius_km: nearbyRadiusKm,
            specialty_id: filters.specialtyId || undefined,
            page,
            limit,
          });

          if (!cancelled) {
            setClinics(res.data || []);
            setTotal(res.meta?.total || 0);
          }
        } else {
          const res = await searchApi.searchClinics({
            search: filters.search.trim() || undefined,
            region: filters.region || undefined,
            district: filters.district || undefined,
            specialty_id: filters.specialtyId || undefined,
            page,
            limit,
          });

          if (!cancelled) {
            setClinics(res.data || []);
            setTotal(res.meta?.total || 0);
          }
        }
      } catch (err: any) {
        if (!cancelled) {
          console.error('Klinikalarni qidirishda xatolik:', err);
          setError('Klinikalarni yuklashda xatolik yuz berdi. Iltimos qayta urinib ko‘ring.');
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    }

    const timer = setTimeout(() => {
      fetchClinics();
    }, 250);

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [filters, page, limit, isNearbyMode, userLocation, nearbyRadiusKm]);

  // Handle filter changes
  const handleFilterChange = (partial: Partial<ClinicFiltersState>) => {
    const updated = { ...filters, ...partial };
    setFilters(updated);
    setPage(1);
    updateUrlParams(updated, 1);
  };

  // Reset filters
  const handleReset = () => {
    const resetState: ClinicFiltersState = {
      search: '',
      region: '',
      district: '',
      specialtyId: '',
    };
    setFilters(resetState);
    setPage(1);
    setIsNearbyMode(false);
    setUserLocation(null);
    updateUrlParams(resetState, 1);
  };

  // Nearby Handlers
  const handleEnableNearby = (coords: { latitude: number; longitude: number }, radius: number) => {
    setUserLocation(coords);
    setNearbyRadiusKm(radius);
    setIsNearbyMode(true);
    setPage(1);
  };

  const handleDisableNearby = () => {
    setIsNearbyMode(false);
    setUserLocation(null);
    setPage(1);
  };

  const totalPages = Math.max(1, Math.ceil(total / limit));

  return (
    <AppLayout>
      <div className="space-y-6">
        {/* Page Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 text-brand-blue text-xs font-bold uppercase tracking-wider mb-1">
              <Building2 className="w-4 h-4" />
              <span>Klinikalar katalogi</span>
            </div>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-brand-navy tracking-tight">
              Klinikalarni qidirish va topish
            </h1>
            <p className="text-sm text-slate-500 mt-1 max-w-2xl">
              O‘zbekistondagi eng yaxshi xususiy klinikalar, tibbiy markazlar va ulardagi tajribali mutaxassislar ro‘yxati.
            </p>
          </div>

          {/* Results Counter Badge */}
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center px-3.5 py-1.5 rounded-full bg-white border border-slate-200/80 text-xs font-semibold text-brand-navy shadow-xs">
              Jami: <strong className="text-brand-blue ml-1 font-bold">{total}</strong> ta klinika
            </span>
          </div>
        </div>

        {/* Filters Bar */}
        <ClinicFilterBar
          filters={filters}
          onFilterChange={handleFilterChange}
          onReset={handleReset}
          regions={regions}
          specialties={specialties}
          isNearbyMode={isNearbyMode}
          onEnableNearby={handleEnableNearby}
          onDisableNearby={handleDisableNearby}
          nearbyRadiusKm={nearbyRadiusKm}
          onRadiusChange={setNearbyRadiusKm}
          viewMode={viewMode}
          onViewModeChange={setViewMode}
          isLoading={isLoading}
        />

        {/* Content Display based on viewMode */}
        {error ? (
          <div className="p-6 bg-red-50 text-red-700 rounded-2xl border border-red-200 text-center">
            <p className="font-semibold">{error}</p>
            <Button variant="outline" size="sm" onClick={() => handleFilterChange({})} className="mt-3">
              Qayta urinish
            </Button>
          </div>
        ) : isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {Array.from({ length: 6 }).map((_, i) => (
              <ClinicCardSkeleton key={i} />
            ))}
          </div>
        ) : clinics.length === 0 ? (
          <EmptySearchState onReset={handleReset} />
        ) : (
          <>
            {/* Split View (Desktop Side-by-side) */}
            {viewMode === 'split' && (
              <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
                {/* Clinics List (Left Column) */}
                <div className="lg:col-span-7 xl:col-span-8 space-y-4">
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    {clinics.map((clinic) => (
                      <ClinicCard
                        key={clinic.id}
                        clinic={clinic}
                        isSelected={selectedClinicId === clinic.id}
                        onSelect={() => setSelectedClinicId(clinic.id)}
                      />
                    ))}
                  </div>

                  {/* Pagination */}
                  {totalPages > 1 && (
                    <div className="pt-6 flex items-center justify-between border-t border-slate-200">
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

                {/* Sticky Interactive Map (Right Column) */}
                <div className="hidden lg:block lg:col-span-5 xl:col-span-4 sticky top-24">
                  <InteractiveMap
                    clinics={clinics}
                    userLocation={userLocation}
                    selectedClinicId={selectedClinicId}
                    onSelectClinic={setSelectedClinicId}
                    height="680px"
                  />
                </div>
              </div>
            )}

            {/* List Only View */}
            {viewMode === 'list' && (
              <div className="space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                  {clinics.map((clinic) => (
                    <ClinicCard
                      key={clinic.id}
                      clinic={clinic}
                      isSelected={selectedClinicId === clinic.id}
                      onSelect={() => setSelectedClinicId(clinic.id)}
                    />
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

            {/* Map View */}
            {viewMode === 'map' && (
              <div className="space-y-4">
                {selectedClinic ? (
                  /* Karta tanlanganda: Desktopda chapda Xarita, o'ngda to'liq ma'lumot (Side-by-Side) */
                  <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
                    {/* Chap ustun: Xarita */}
                    <div className="lg:col-span-7 xl:col-span-8 space-y-3">
                      <InteractiveMap
                        clinics={clinics}
                        userLocation={userLocation}
                        selectedClinicId={selectedClinicId}
                        onSelectClinic={(id) => setSelectedClinicId(id)}
                        height="680px"
                      />

                      {/* Xarita ostidagi tezkor klinika tanlash tugmalari */}
                      <div className="bg-white rounded-2xl p-3 border border-slate-200/80 shadow-xs">
                        <div className="flex items-center justify-between mb-2 px-1">
                          <span className="text-xs font-bold text-slate-500 uppercase tracking-wider">
                            Xaritadagi klinikalar ({clinics.length}):
                          </span>
                          <button
                            type="button"
                            onClick={() => setSelectedClinicId(null)}
                            className="text-[11px] text-brand-blue hover:underline font-semibold"
                          >
                            Xaritani to‘liq ochish
                          </button>
                        </div>
                        <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-thin">
                          {clinics.map((c) => {
                            const isSel = selectedClinic?.id === c.id;
                            return (
                              <button
                                key={c.id}
                                type="button"
                                onClick={() => setSelectedClinicId(c.id)}
                                className={`shrink-0 flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-semibold border transition-all ${
                                  isSel
                                    ? 'bg-brand-navy text-white border-brand-navy shadow-xs ring-2 ring-brand-blue/30 scale-102'
                                    : 'bg-slate-50 text-slate-700 border-slate-200 hover:bg-slate-100'
                                }`}
                              >
                                <span className="w-2.5 h-2.5 rounded-full bg-red-500 shrink-0"></span>
                                <span className="truncate max-w-[150px]">{c.name}</span>
                                {c.region && (
                                  <span
                                    className={`text-[10px] ${isSel ? 'text-slate-300' : 'text-slate-400'}`}
                                  >
                                    ({c.region})
                                  </span>
                                )}
                              </button>
                            );
                          })}
                        </div>
                      </div>
                    </div>

                    {/* O'ng ustun (Desktopda): Tanlangan klinikaning to'liq ma'lumot kartasi */}
                    <div className="hidden lg:block lg:col-span-5 xl:col-span-4 lg:sticky lg:top-24">
                      <ClinicMapDetailCard
                        clinic={selectedClinic}
                        onClose={() => setSelectedClinicId(null)}
                      />
                    </div>

                    {/* Mobilda: Pastdan chiquvchi modal (Bottom Sheet Drawer) */}
                    <div className="lg:hidden fixed inset-x-0 bottom-0 z-50 p-3 max-h-[85vh] overflow-y-auto bg-black/40 backdrop-blur-xs flex flex-col justify-end">
                      <div className="max-w-md mx-auto w-full">
                        <ClinicMapDetailCard
                          clinic={selectedClinic}
                          onClose={() => setSelectedClinicId(null)}
                        />
                      </div>
                    </div>
                  </div>
                ) : (
                  /* Dastlabki holat: Xarita 100% to'liq kenglikni egallaydi */
                  <div className="space-y-3">
                    <div className="relative w-full">
                      <InteractiveMap
                        clinics={clinics}
                        userLocation={userLocation}
                        selectedClinicId={selectedClinicId}
                        onSelectClinic={(id) => setSelectedClinicId(id)}
                        height="700px"
                      />

                      {/* Xarita ichida yo'riqnoma banner */}
                      <div className="absolute top-4 left-1/2 -translate-x-1/2 z-10 pointer-events-none">
                        <div className="bg-white/95 backdrop-blur-md px-4 py-2 rounded-full border border-slate-200 shadow-elevated flex items-center gap-2 text-xs font-bold text-brand-navy">
                          <span className="w-3 h-3 rounded-full bg-red-500 animate-pulse"></span>
                          <span>Barcha klinikalar ko‘rsatilgan — to‘liq ma’lumot olish uchun istalgan qizil pinni bosing</span>
                        </div>
                      </div>
                    </div>

                    {/* Xarita ostidagi klinikalar ro'yxati */}
                    <div className="bg-white rounded-2xl p-3.5 border border-slate-200/80 shadow-xs">
                      <div className="flex items-center justify-between mb-2 px-1">
                        <span className="text-xs font-bold text-slate-500 uppercase tracking-wider">
                          Xaritadagi tasdiqlangan klinikalar ({clinics.length}):
                        </span>
                        <span className="text-[11px] text-slate-400">
                          Klinikani bosib uning ma’lumotlarini o‘ngda ochishingiz mumkin
                        </span>
                      </div>
                      <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-thin">
                        {clinics.map((c) => (
                          <button
                            key={c.id}
                            type="button"
                            onClick={() => setSelectedClinicId(c.id)}
                            className="shrink-0 flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-semibold border border-slate-200 bg-slate-50 text-slate-700 hover:bg-white hover:border-brand-blue hover:text-brand-blue transition-all"
                          >
                            <span className="w-2.5 h-2.5 rounded-full bg-red-500 shrink-0"></span>
                            <span className="truncate max-w-[150px]">{c.name}</span>
                            {c.region && (
                              <span className="text-[10px] text-slate-400">({c.region})</span>
                            )}
                          </button>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )}
          </>
        )}
      </div>
    </AppLayout>
  );
}

