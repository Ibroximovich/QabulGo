'use client';

import React, { useEffect, useState } from 'react';
import { Search, MapPin, Stethoscope, RotateCcw, Navigation, List, Map as MapIcon, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { RegionItem, SpecialtyItem } from '../types';
import { toast } from 'sonner';

export interface ClinicFiltersState {
  search: string;
  region: string;
  district: string;
  specialtyId: string;
}

interface ClinicFilterBarProps {
  filters: ClinicFiltersState;
  onFilterChange: (filters: Partial<ClinicFiltersState>) => void;
  onReset: () => void;
  regions: RegionItem[];
  specialties: SpecialtyItem[];
  // Nearby Mode
  isNearbyMode: boolean;
  onEnableNearby: (coords: { latitude: number; longitude: number }, radiusKm: number) => void;
  onDisableNearby: () => void;
  nearbyRadiusKm: number;
  onRadiusChange: (radiusKm: number) => void;
  // View mode
  viewMode: 'list' | 'map' | 'split';
  onViewModeChange: (mode: 'list' | 'map' | 'split') => void;
  isLoading?: boolean;
}

export function ClinicFilterBar({
  filters,
  onFilterChange,
  onReset,
  regions,
  specialties,
  isNearbyMode,
  onEnableNearby,
  onDisableNearby,
  nearbyRadiusKm,
  onRadiusChange,
  viewMode,
  onViewModeChange,
  isLoading = false,
}: ClinicFilterBarProps) {
  const [isLocating, setIsLocating] = useState(false);

  // Available districts based on selected region
  const selectedRegionObj = regions.find((r) => r.region === filters.region);
  const availableDistricts = selectedRegionObj?.districts || [];

  const handleGeolocationClick = () => {
    if (!navigator.geolocation) {
      toast.error('Brauzeringiz geolokatsiyani qo‘llab-quvvatlamaydi');
      return;
    }

    setIsLocating(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setIsLocating(false);
        const { latitude, longitude } = position.coords;
        toast.success('Joylashuvingiz aniqlandi!');
        onEnableNearby({ latitude, longitude }, nearbyRadiusKm);
      },
      (error) => {
        setIsLocating(false);
        if (error.code === error.PERMISSION_DENIED) {
          toast.error('Geolokatsiyaga ruxsat berilmadi. Oddiy hudud filtri orqali qidirishingiz mumkin.');
        } else if (error.code === error.POSITION_UNAVAILABLE) {
          toast.error('Joylashuv ma’lumoti mavjud emas.');
        } else {
          toast.error('Geolokatsiyani aniqlashda xatolik yuz berdi.');
        }
      },
      { timeout: 10000, enableHighAccuracy: true }
    );
  };

  const hasActiveFilters = Boolean(
    filters.search || filters.region || filters.district || filters.specialtyId || isNearbyMode
  );

  return (
    <div className="bg-white rounded-3xl border border-slate-200/80 p-4 sm:p-6 shadow-subtle space-y-4">
      {/* Top Row: Search input + Nearby Button + View Mode Toggle */}
      <div className="flex flex-col lg:flex-row items-stretch lg:items-center justify-between gap-3">
        {/* Search bar */}
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
          <input
            type="text"
            value={filters.search}
            onChange={(e) => onFilterChange({ search: e.target.value })}
            placeholder="Klinika nomi, manzili yoki tavsifi bo‘yicha qidirish..."
            className="w-full pl-11 pr-4 py-3 rounded-2xl bg-slate-50 border border-slate-200 text-sm text-brand-navy placeholder-slate-400 focus:bg-white focus:border-brand-blue focus:ring-2 focus:ring-brand-blue/20 transition-all outline-none"
          />
          {isLoading && (
            <div className="absolute right-4 top-1/2 -translate-y-1/2">
              <Loader2 className="w-4 h-4 text-brand-blue animate-spin" />
            </div>
          )}
        </div>

        {/* Nearby Clinics Button */}
        <div className="flex items-center gap-2">
          {isNearbyMode ? (
            <div className="flex items-center gap-2 bg-teal-50 border border-teal-200 px-3 py-1.5 rounded-2xl">
              <span className="text-xs font-bold text-brand-teal flex items-center gap-1.5">
                <span className="w-2 h-2 rounded-full bg-brand-teal animate-ping" />
                <span>Yaqin klinikalar:</span>
              </span>
              <select
                value={nearbyRadiusKm}
                onChange={(e) => onRadiusChange(Number(e.target.value))}
                className="bg-white text-xs font-semibold text-brand-navy py-1 px-2 rounded-lg border border-teal-200 outline-none focus:ring-1 focus:ring-brand-teal"
              >
                <option value={5}>5 km</option>
                <option value={10}>10 km</option>
                <option value={20}>20 km</option>
                <option value={50}>50 km</option>
              </select>
              <button
                type="button"
                onClick={onDisableNearby}
                className="text-xs text-slate-500 hover:text-red-600 font-medium ml-1"
                title="Yaqin klinikalar rejimidan chiqish"
              >
                ✕
              </button>
            </div>
          ) : (
            <Button
              type="button"
              variant="outline"
              onClick={handleGeolocationClick}
              disabled={isLocating}
              className="gap-2 shrink-0 border-brand-teal/40 text-brand-teal hover:bg-brand-teal/10 hover:text-brand-teal"
            >
              {isLocating ? (
                <Loader2 className="w-4 h-4 animate-spin text-brand-teal" />
              ) : (
                <Navigation className="w-4 h-4 text-brand-teal" />
              )}
              <span>📍 Yaqin klinikalarni topish</span>
            </Button>
          )}

          {/* View Mode Toggle (Desktop & Mobile) */}
          <div className="hidden sm:flex items-center bg-slate-100 p-1 rounded-2xl border border-slate-200/80">
            <button
              type="button"
              onClick={() => onViewModeChange('list')}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
                viewMode === 'list'
                  ? 'bg-white text-brand-navy shadow-xs'
                  : 'text-slate-600 hover:text-brand-navy'
              }`}
            >
              <List className="w-3.5 h-3.5" />
              <span>Ro‘yxat</span>
            </button>
            <button
              type="button"
              onClick={() => onViewModeChange('split')}
              className={`hidden lg:flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
                viewMode === 'split'
                  ? 'bg-white text-brand-navy shadow-xs'
                  : 'text-slate-600 hover:text-brand-navy'
              }`}
            >
              <span className="w-3.5 h-3.5 border-r border-slate-300 pr-1 inline-block">◫</span>
              <span>Yonma-yon</span>
            </button>
            <button
              type="button"
              onClick={() => onViewModeChange('map')}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
                viewMode === 'map'
                  ? 'bg-white text-brand-navy shadow-xs'
                  : 'text-slate-600 hover:text-brand-navy'
              }`}
            >
              <MapIcon className="w-3.5 h-3.5" />
              <span>Xarita</span>
            </button>
          </div>
        </div>
      </div>

      {/* Bottom Row: Region, District, Specialty, Reset */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 pt-2 border-t border-slate-100">
        {/* Region */}
        <div>
          <label className="block text-[11px] font-semibold text-slate-500 mb-1">Viloyat</label>
          <div className="relative">
            <MapPin className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
            <select
              value={filters.region}
              onChange={(e) => onFilterChange({ region: e.target.value, district: '' })}
              className="w-full pl-9 pr-8 py-2.5 rounded-xl bg-slate-50 border border-slate-200 text-xs font-medium text-brand-navy focus:bg-white focus:border-brand-blue outline-none"
            >
              <option value="">Barcha viloyatlar</option>
              {regions.map((r) => (
                <option key={r.region} value={r.region}>
                  {r.region}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* District */}
        <div>
          <label className="block text-[11px] font-semibold text-slate-500 mb-1">Tuman / Shahar</label>
          <select
            value={filters.district}
            onChange={(e) => onFilterChange({ district: e.target.value })}
            disabled={!filters.region || availableDistricts.length === 0}
            className="w-full px-3 py-2.5 rounded-xl bg-slate-50 border border-slate-200 text-xs font-medium text-brand-navy focus:bg-white focus:border-brand-blue outline-none disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <option value="">Barcha tumanlar</option>
            {availableDistricts.map((d) => (
              <option key={d} value={d}>
                {d}
              </option>
            ))}
          </select>
        </div>

        {/* Specialty */}
        <div>
          <label className="block text-[11px] font-semibold text-slate-500 mb-1">Mutaxassislik</label>
          <div className="relative">
            <Stethoscope className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
            <select
              value={filters.specialtyId}
              onChange={(e) => onFilterChange({ specialtyId: e.target.value })}
              className="w-full pl-9 pr-8 py-2.5 rounded-xl bg-slate-50 border border-slate-200 text-xs font-medium text-brand-navy focus:bg-white focus:border-brand-blue outline-none"
            >
              <option value="">Barcha yo‘nalishlar</option>
              {specialties.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Reset / Actions */}
        <div className="flex items-end">
          <Button
            type="button"
            variant="ghost"
            onClick={onReset}
            disabled={!hasActiveFilters}
            className="w-full gap-2 text-xs text-slate-500 hover:text-brand-navy disabled:opacity-40"
          >
            <RotateCcw className="w-3.5 h-3.5" />
            <span>Filtrlarni tozalash</span>
          </Button>
        </div>
      </div>

      {/* Mobile view switch bar */}
      <div className="flex sm:hidden items-center justify-around bg-slate-100 p-1 rounded-2xl border border-slate-200/80">
        <button
          type="button"
          onClick={() => onViewModeChange('list')}
          className={`flex-1 flex items-center justify-center gap-1.5 py-2 rounded-xl text-xs font-bold transition-all ${
            viewMode === 'list' ? 'bg-white text-brand-navy shadow-xs' : 'text-slate-600'
          }`}
        >
          <List className="w-4 h-4" />
          <span>Ro‘yxat</span>
        </button>
        <button
          type="button"
          onClick={() => onViewModeChange('map')}
          className={`flex-1 flex items-center justify-center gap-1.5 py-2 rounded-xl text-xs font-bold transition-all ${
            viewMode === 'map' ? 'bg-white text-brand-navy shadow-xs' : 'text-slate-600'
          }`}
        >
          <MapIcon className="w-4 h-4" />
          <span>Xarita</span>
        </button>
      </div>
    </div>
  );
}

