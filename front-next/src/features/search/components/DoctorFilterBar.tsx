'use client';

import React from 'react';
import { Search, MapPin, Stethoscope, ArrowUpDown, RotateCcw, DollarSign, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { RegionItem, SpecialtyItem } from '../types';

export interface DoctorFiltersState {
  search: string;
  region: string;
  district: string;
  specialtyId: string;
  minPrice: string;
  maxPrice: string;
  sort: 'name_asc' | 'experience_desc' | 'price_asc' | 'price_desc';
}

interface DoctorFilterBarProps {
  filters: DoctorFiltersState;
  onFilterChange: (filters: Partial<DoctorFiltersState>) => void;
  onReset: () => void;
  regions: RegionItem[];
  specialties: SpecialtyItem[];
  isLoading?: boolean;
}

export function DoctorFilterBar({
  filters,
  onFilterChange,
  onReset,
  regions,
  specialties,
  isLoading = false,
}: DoctorFilterBarProps) {
  const selectedRegionObj = regions.find((r) => r.region === filters.region);
  const availableDistricts = selectedRegionObj?.districts || [];

  const hasActiveFilters = Boolean(
    filters.search ||
      filters.region ||
      filters.district ||
      filters.specialtyId ||
      filters.minPrice ||
      filters.maxPrice ||
      filters.sort !== 'name_asc'
  );

  return (
    <div className="bg-white rounded-3xl border border-slate-200/80 p-4 sm:p-6 shadow-subtle space-y-4">
      {/* Top Row: Search and Sort */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        {/* Search Input */}
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
          <input
            type="text"
            value={filters.search}
            onChange={(e) => onFilterChange({ search: e.target.value })}
            placeholder="Shifokor ismi, mutaxassisligi yoki ta’rifi bo‘yicha qidirish..."
            className="w-full pl-11 pr-4 py-3 rounded-2xl bg-slate-50 border border-slate-200 text-sm text-brand-navy placeholder-slate-400 focus:bg-white focus:border-brand-blue focus:ring-2 focus:ring-brand-blue/20 transition-all outline-none"
          />
          {isLoading && (
            <div className="absolute right-4 top-1/2 -translate-y-1/2">
              <Loader2 className="w-4 h-4 text-brand-blue animate-spin" />
            </div>
          )}
        </div>

        {/* Sorting Dropdown */}
        <div className="flex items-center gap-2 shrink-0">
          <div className="relative">
            <ArrowUpDown className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
            <select
              value={filters.sort}
              onChange={(e) => onFilterChange({ sort: e.target.value as any })}
              className="pl-9 pr-8 py-3 rounded-2xl bg-slate-50 border border-slate-200 text-xs font-semibold text-brand-navy focus:bg-white focus:border-brand-blue outline-none"
            >
              <option value="name_asc">Alifbo bo‘yicha (A-Z)</option>
              <option value="experience_desc">Ko‘p yillik tajriba</option>
              <option value="price_asc">Arzonroq narx</option>
              <option value="price_desc">Qimmatroq narx</option>
            </select>
          </div>
        </div>
      </div>

      {/* Bottom Grid: Region, District, Specialty, Price Range, Reset */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3 pt-2 border-t border-slate-100">
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
              <option value="">Barcha mutaxassisliklar</option>
              {specialties.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Price Range */}
        <div>
          <label className="block text-[11px] font-semibold text-slate-500 mb-1">Narx oralig‘i (so‘m)</label>
          <div className="flex items-center gap-1.5">
            <input
              type="number"
              value={filters.minPrice}
              onChange={(e) => onFilterChange({ minPrice: e.target.value })}
              placeholder="Dan"
              className="w-1/2 px-2.5 py-2.5 rounded-xl bg-slate-50 border border-slate-200 text-xs font-medium text-brand-navy focus:bg-white focus:border-brand-blue outline-none"
            />
            <span className="text-slate-300">-</span>
            <input
              type="number"
              value={filters.maxPrice}
              onChange={(e) => onFilterChange({ maxPrice: e.target.value })}
              placeholder="Gacha"
              className="w-1/2 px-2.5 py-2.5 rounded-xl bg-slate-50 border border-slate-200 text-xs font-medium text-brand-navy focus:bg-white focus:border-brand-blue outline-none"
            />
          </div>
        </div>

        {/* Reset Actions */}
        <div className="flex items-end">
          <Button
            type="button"
            variant="ghost"
            onClick={onReset}
            disabled={!hasActiveFilters}
            className="w-full gap-2 text-xs text-slate-500 hover:text-brand-navy disabled:opacity-40"
          >
            <RotateCcw className="w-3.5 h-3.5" />
            <span>Tozalash</span>
          </Button>
        </div>
      </div>
    </div>
  );
}

