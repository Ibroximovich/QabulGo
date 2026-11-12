'use client';

import React, { useEffect, useRef, useState, useCallback } from 'react';
import Link from 'next/link';
import 'leaflet/dist/leaflet.css';
import { MapPin } from 'lucide-react';
import { extractCoordinatesFromUrl } from '@/lib/utils/location';
import { PublicClinicItem, NearbyClinicItem } from '@/features/search/types';

export interface MapClinicItem {
  id: string;
  name: string;
  address: string;
  region?: string;
  district?: string;
  latitude?: number | null;
  longitude?: number | null;
  location_url?: string;
  distance_km?: number;
  phone?: string;
  logo_url?: string;
  doctors_count?: number;
}

export interface InteractiveMapProps {
  clinics?: (PublicClinicItem | NearbyClinicItem | MapClinicItem)[];
  userLocation?: { latitude: number; longitude: number } | null;
  selectedClinicId?: string | null;
  onSelectClinic?: (clinicId: string) => void;
  // Picker mode for Clinic Administrator:
  pickerMode?: boolean;
  pickedLocation?: { latitude: number; longitude: number } | null;
  onLocationPick?: (coords: { latitude: number; longitude: number }) => void;
  height?: string;
  zoom?: number;
  center?: [number, number];
}

// O'zbekiston / Toshkent markazi (default)
const DEFAULT_CENTER: [number, number] = [41.311081, 69.240562];

interface MarkerEntry {
  marker: any;
  clinic: any;
  lat: number;
  lng: number;
}

export function InteractiveMap({
  clinics = [],
  userLocation = null,
  selectedClinicId = null,
  onSelectClinic,
  pickerMode = false,
  pickedLocation = null,
  onLocationPick,
  height = '500px',
  zoom = 12,
  center = DEFAULT_CENTER,
}: InteractiveMapProps) {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapInstanceRef = useRef<any>(null);
  const markersLayerRef = useRef<any>(null);
  const markersMapRef = useRef<Map<string, MarkerEntry>>(new Map());
  const [isMounted, setIsMounted] = useState(false);
  const [mapReady, setMapReady] = useState(false);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  // 1. Initialize Leaflet Map
  useEffect(() => {
    if (!isMounted || !mapContainerRef.current) return;

    let isCancelled = false;

    async function initMap() {
      const L = await import('leaflet');

      if (isCancelled || !mapContainerRef.current) return;

      // Agar avvalgi xarita bo'lsa, tozalash
      if (mapInstanceRef.current) {
        mapInstanceRef.current.remove();
        mapInstanceRef.current = null;
      }

      // Initial center hisoblash
      let initialCenter = center;
      if (pickedLocation) {
        initialCenter = [pickedLocation.latitude, pickedLocation.longitude];
      } else if (userLocation) {
        initialCenter = [userLocation.latitude, userLocation.longitude];
      } else if (clinics.length > 0) {
        const firstWithCoords = clinics.find((c) => {
          const lat = typeof c.latitude === 'number' && !isNaN(c.latitude) && c.latitude !== 0;
          const lon = typeof c.longitude === 'number' && !isNaN(c.longitude) && c.longitude !== 0;
          return lat && lon;
        });
        if (firstWithCoords?.latitude && firstWithCoords?.longitude) {
          initialCenter = [firstWithCoords.latitude, firstWithCoords.longitude];
        }
      }

      const map = L.map(mapContainerRef.current, {
        center: initialCenter,
        zoom,
        zoomControl: true,
        scrollWheelZoom: true,
      });

      // OpenStreetMap Tile Layer
      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution:
          '&copy; <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener">OpenStreetMap</a>',
        maxZoom: 19,
      }).addTo(map);

      mapInstanceRef.current = map;
      markersLayerRef.current = L.layerGroup().addTo(map);
      setMapReady(true);

      // Picker mode click handler
      if (pickerMode) {
        map.on('click', (e: any) => {
          const lat = parseFloat(e.latlng.lat.toFixed(6));
          const lng = parseFloat(e.latlng.lng.toFixed(6));
          if (onLocationPick) {
            onLocationPick({ latitude: lat, longitude: lng });
          }
        });
      }
    }

    initMap();

    // Map container o'lchami o'zgarganda xaritani moslash
    let resizeObserver: ResizeObserver | null = null;
    if (typeof ResizeObserver !== 'undefined' && mapContainerRef.current) {
      resizeObserver = new ResizeObserver(() => {
        if (mapInstanceRef.current) {
          mapInstanceRef.current.invalidateSize();
        }
      });
      resizeObserver.observe(mapContainerRef.current);
    }

    return () => {
      isCancelled = true;
      if (resizeObserver) {
        resizeObserver.disconnect();
      }
      if (mapInstanceRef.current) {
        mapInstanceRef.current.remove();
        mapInstanceRef.current = null;
      }
      setMapReady(false);
    };
  }, [isMounted]);

  // Helper: Icon yaratish
  const createClinicIcon = useCallback((L: any, clinicName: string, isSelected: boolean) => {
    const redPinSvg = isSelected
      ? `<svg width="38" height="48" viewBox="0 0 34 44" fill="none" xmlns="http://www.w3.org/2000/svg" style="filter: drop-shadow(0 4px 8px rgba(220,38,38,0.55));">
           <path d="M17 0C7.611 0 0 7.611 0 17C0 29.75 15.3 43.15 15.95 43.71C16.25 43.97 16.62 44.1 17 44.1C17.38 44.1 17.75 43.97 18.05 43.71C18.7 43.15 34 29.75 34 17C34 7.611 26.389 0 17 0Z" fill="#DC2626" stroke="#FFFFFF" stroke-width="2.5"/>
           <circle cx="17" cy="16" r="6.5" fill="#FFFFFF"/>
         </svg>`
      : `<svg width="32" height="42" viewBox="0 0 34 44" fill="none" xmlns="http://www.w3.org/2000/svg" style="filter: drop-shadow(0 3px 4px rgba(0,0,0,0.35));">
           <path d="M17 0C7.611 0 0 7.611 0 17C0 29.75 15.3 43.15 15.95 43.71C16.25 43.97 16.62 44.1 17 44.1C17.38 44.1 17.75 43.97 18.05 43.71C18.7 43.15 34 29.75 34 17C34 7.611 26.389 0 17 0Z" fill="#EA4335" stroke="#B91C1C" stroke-width="1.2"/>
           <circle cx="17" cy="16" r="5.5" fill="#FFFFFF"/>
         </svg>`;

    return L.divIcon({
      className: 'qabulgo-red-pin',
      html: `
        <div style="position: relative; display: flex; flex-direction: column; align-items: center; justify-content: center; cursor: pointer; transform: ${
          isSelected ? 'scale(1.22)' : 'scale(1)'
        }; transition: transform 0.2s ease;">
          ${redPinSvg}
          <div style="width: 14px; height: 5px; background: rgba(0,0,0,0.25); border-radius: 50%; filter: blur(1px); margin-top: -2px;"></div>
          ${
            isSelected
              ? `<div style="position: absolute; top: -30px; background: #08234A; color: white; font-size: 11px; font-weight: bold; padding: 3px 9px; border-radius: 8px; white-space: nowrap; box-shadow: 0 4px 8px rgba(0,0,0,0.35); pointer-events: none; border: 1px solid rgba(255,255,255,0.2);">${clinicName}</div>`
              : ''
          }
        </div>
      `,
      iconSize: isSelected ? [38, 48] : [32, 42],
      iconAnchor: isSelected ? [19, 48] : [16, 42],
      popupAnchor: isSelected ? [0, -48] : [0, -42],
    });
  }, []);

  // Helper: Popup kontentini yaratish
  const createPopupContent = useCallback((clinic: any) => {
    const distanceStr =
      'distance_km' in clinic && typeof clinic.distance_km === 'number'
        ? `<div class="inline-flex items-center gap-1 text-[11px] font-semibold text-brand-teal bg-teal-50 px-2 py-0.5 rounded-md border border-teal-100 my-1">
             <span>📍 ${clinic.distance_km.toFixed(1)} km masofada</span>
           </div>`
        : '';

    const phoneStr = clinic.phone
      ? `<div class="mt-1.5 flex items-center gap-1.5 text-xs text-slate-600 font-medium">
           <span>📞</span>
           <a href="tel:${clinic.phone}" class="text-brand-blue hover:underline font-semibold">${clinic.phone}</a>
         </div>`
      : '';

    return `
      <div class="p-2.5 font-sans min-w-[210px] max-w-[250px]">
        <h4 class="font-bold text-sm text-brand-navy leading-snug line-clamp-1">${clinic.name}</h4>
        <p class="text-xs text-slate-500 mt-1 line-clamp-2">${clinic.address || clinic.district || clinic.region || ''}</p>
        ${distanceStr}
        ${phoneStr}
        <div class="mt-3 pt-2.5 border-t border-slate-100 flex items-center justify-between gap-2">
          <a href="/clinics/${clinic.id}" class="inline-flex items-center justify-center px-3 py-1.5 text-xs font-semibold rounded-lg bg-brand-blue text-white hover:bg-brand-blue-600 transition-colors shadow-xs w-full text-center">
            Klinikani ko‘rish &rarr;
          </a>
        </div>
      </div>
    `;
  }, []);

  // 2. Render Markers Layer (Only rebuilds when clinics list or user/picker location changes)
  useEffect(() => {
    if (!mapReady || !mapInstanceRef.current || !markersLayerRef.current) return;

    let isCancelled = false;

    async function buildLayers() {
      const L = await import('leaflet');
      if (isCancelled || !mapInstanceRef.current || !markersLayerRef.current) return;

      const map = mapInstanceRef.current;
      const markersLayer = markersLayerRef.current;
      markersLayer.clearLayers();
      markersMapRef.current.clear();

      // Haqiqiy koordinataga ega klinikalarni saralash (Termiz fallback yo'q!)
      const validClinicsWithCoords = clinics
        .map((c) => {
          let lat =
            typeof c.latitude === 'number' && !isNaN(c.latitude) && c.latitude !== 0 ? c.latitude : null;
          let lon =
            typeof c.longitude === 'number' && !isNaN(c.longitude) && c.longitude !== 0 ? c.longitude : null;

          if (lat === null || lon === null) {
            const parsed = extractCoordinatesFromUrl((c as any).location_url);
            if (parsed) {
              lat = parsed.latitude;
              lon = parsed.longitude;
            }
          }

          if (lat === null || lon === null) {
            return null;
          }

          return {
            ...c,
            latitude: lat,
            longitude: lon,
          };
        })
        .filter((c): c is (typeof clinics[0] & { latitude: number; longitude: number }) => c !== null);

      // Spiderfy / Bir xil koordinatadagi klinikalarni tarqatish
      // Bir-biriga o'ta yaqin (< ~25 metr) bo'lgan klinikalarni guruhlash
      const coordGroups = new Map<string, typeof validClinicsWithCoords>();
      validClinicsWithCoords.forEach((clinic) => {
        const key = `${clinic.latitude.toFixed(4)}_${clinic.longitude.toFixed(4)}`;
        if (!coordGroups.has(key)) {
          coordGroups.set(key, []);
        }
        coordGroups.get(key)!.push(clinic);
      });

      interface PositionedItem {
        clinic: typeof validClinicsWithCoords[0];
        lat: number;
        lng: number;
      }

      const positionedItems: PositionedItem[] = [];
      coordGroups.forEach((group) => {
        const count = group.length;
        if (count === 1) {
          positionedItems.push({
            clinic: group[0],
            lat: group[0].latitude,
            lng: group[0].longitude,
          });
        } else {
          // Aylana bo'ylab tarqatish (spiderfy offset)
          const radius = 0.00028; // ~30 metr
          group.forEach((clinic, idx) => {
            const angle = (2 * Math.PI / count) * idx;
            const offsetLat = clinic.latitude + radius * Math.cos(angle);
            const offsetLng =
              clinic.longitude + (radius * Math.sin(angle)) / Math.cos((clinic.latitude * Math.PI) / 180);
            positionedItems.push({
              clinic,
              lat: offsetLat,
              lng: offsetLng,
            });
          });
        }
      });

      const bounds = L.latLngBounds([]);

      // 1. Foydalanuvchi joylashuvi marker
      if (userLocation && typeof userLocation.latitude === 'number' && typeof userLocation.longitude === 'number') {
        const userIcon = L.divIcon({
          className: 'custom-user-marker',
          html: `
            <div class="relative flex items-center justify-center w-8 h-8">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-brand-blue opacity-60"></span>
              <span class="relative inline-flex rounded-full h-4 w-4 bg-brand-blue border-2 border-white shadow-md"></span>
            </div>
          `,
          iconSize: [32, 32],
          iconAnchor: [16, 16],
        });

        const uMarker = L.marker([userLocation.latitude, userLocation.longitude], {
          icon: userIcon,
          title: 'Sizning joylashuvingiz',
        }).bindPopup(`
          <div class="p-1 font-sans text-xs">
            <p class="font-bold text-brand-navy">📍 Sizning joylashuvingiz</p>
            <p class="text-[11px] text-slate-500 mt-0.5">${userLocation.latitude.toFixed(4)}, ${userLocation.longitude.toFixed(4)}</p>
          </div>
        `);

        markersLayer.addLayer(uMarker);
        bounds.extend([userLocation.latitude, userLocation.longitude]);
      }

      // 2. Picker rejimi
      if (pickerMode && pickedLocation) {
        const pickerIcon = L.divIcon({
          className: 'custom-picker-marker',
          html: `
            <div class="w-9 h-9 rounded-full bg-emerald-600 text-white shadow-elevated border-2 border-white flex items-center justify-center animate-bounce">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>
            </div>
          `,
          iconSize: [36, 36],
          iconAnchor: [18, 36],
        });

        const pMarker = L.marker([pickedLocation.latitude, pickedLocation.longitude], {
          icon: pickerIcon,
          draggable: true,
          title: 'Klinika lokatsiyasi',
        });

        pMarker.on('dragend', (e: any) => {
          const latlng = e.target.getLatLng();
          if (onLocationPick) {
            onLocationPick({
              latitude: parseFloat(latlng.lat.toFixed(6)),
              longitude: parseFloat(latlng.lng.toFixed(6)),
            });
          }
        });

        pMarker.bindPopup(`
          <div class="p-1 font-sans text-xs">
            <p class="font-bold text-emerald-700">📍 Tanlangan klinika koordinatalari</p>
            <p class="text-[11px] text-slate-600 mt-1">${pickedLocation.latitude.toFixed(6)}, ${pickedLocation.longitude.toFixed(6)}</p>
            <p class="text-[10px] text-slate-400 mt-0.5">Surib joyini o'zgartirishingiz mumkin</p>
          </div>
        `);

        markersLayer.addLayer(pMarker);
        bounds.extend([pickedLocation.latitude, pickedLocation.longitude]);
      }

      // 3. Klinikalar markerlari
      positionedItems.forEach(({ clinic, lat, lng }) => {
        const isSelected = selectedClinicId === clinic.id;
        const icon = createClinicIcon(L, clinic.name, isSelected);

        const marker = L.marker([lat, lng], {
          icon,
          title: clinic.name,
          zIndexOffset: isSelected ? 1000 : 0,
        });

        const popupContent = createPopupContent(clinic);
        marker.bindPopup(popupContent);

        marker.on('click', () => {
          if (onSelectClinic) {
            onSelectClinic(clinic.id);
          }
          map.panTo([lat, lng], { animate: true, duration: 0.5 });
        });

        markersLayer.addLayer(marker);
        markersMapRef.current.set(clinic.id, { marker, clinic, lat, lng });
        bounds.extend([lat, lng]);

        if (isSelected) {
          marker.openPopup();
        }
      });

      // Sahifa yuklanganda / natijalar o'zgarganda fitBounds
      if (bounds.isValid() && positionedItems.length > 1) {
        map.fitBounds(bounds, { padding: [50, 50], maxZoom: 15 });
      } else if (positionedItems.length === 1) {
        map.setView([positionedItems[0].lat, positionedItems[0].lng], 14);
      }
    }

    buildLayers();

    return () => {
      isCancelled = true;
    };
  }, [mapReady, clinics, userLocation, pickedLocation, pickerMode, createClinicIcon, createPopupContent, onSelectClinic, onLocationPick]);

  // 3. Handle selectedClinicId changes (pans, highlights marker, opens popup WITHOUT resetting layers)
  useEffect(() => {
    if (!mapReady || !mapInstanceRef.current) return;

    let isCancelled = false;

    async function updateSelection() {
      const L = await import('leaflet');
      if (isCancelled || !mapInstanceRef.current) return;

      const map = mapInstanceRef.current;
      const markersMap = markersMapRef.current;

      markersMap.forEach((entry, clinicId) => {
        const isSelected = selectedClinicId === clinicId;
        const icon = createClinicIcon(L, entry.clinic.name, isSelected);
        entry.marker.setIcon(icon);
        entry.marker.setZIndexOffset(isSelected ? 1000 : 0);

        if (isSelected) {
          entry.marker.openPopup();
          // Xaritani tanlangan klinikaga markazlashtirish
          map.panTo([entry.lat, entry.lng], { animate: true, duration: 0.5 });
        }
      });
    }

    updateSelection();

    return () => {
      isCancelled = true;
    };
  }, [selectedClinicId, mapReady, createClinicIcon]);

  return (
    <div className="relative w-full rounded-2xl overflow-hidden border border-slate-200/80 shadow-subtle bg-slate-100">
      <div ref={mapContainerRef} style={{ height, width: '100%' }} className="z-0" />

      {/* Attribution badge overlay */}
      <div className="absolute bottom-2 left-2 z-10 pointer-events-none">
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-white/90 backdrop-blur-sm text-[11px] font-medium text-slate-600 border border-slate-200/60 shadow-xs">
          <MapPin className="w-3.5 h-3.5 text-brand-blue" />
          <span>QabulGo Xaritasi</span>
        </span>
      </div>
    </div>
  );
}
