'use client';

import React, { useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import 'leaflet/dist/leaflet.css';
import { Building2, Navigation, MapPin } from 'lucide-react';
import { extractCoordinatesFromUrl, REGION_DEFAULT_COORDINATES } from '@/lib/utils/location';
import { PublicClinicItem, NearbyClinicItem } from '@/features/search/types';

export interface MapClinicItem {
  id: string;
  name: string;
  address: string;
  region?: string;
  district?: string;
  latitude: number;
  longitude: number;
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
  const userMarkerRef = useRef<any>(null);
  const pickerMarkerRef = useRef<any>(null);
  const [isMounted, setIsMounted] = useState(false);
  const [mapInstance, setMapInstance] = useState<any>(null);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  // Initialize Map
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
        const firstWithCoords = clinics.find((c) => c.latitude && c.longitude);
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
      setMapInstance(map);

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

    // Map container o'lchami o'zgarganda (masalan, yon panel ochilganda/yopilganda) xaritani moslash
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
    };
  }, [isMounted]);

  // Update Markers (Clinics, User, Picker)
  useEffect(() => {
    if (!mapInstance || !markersLayerRef.current) return;

    let isCancelled = false;

    async function renderLayers() {
      const L = await import('leaflet');
      if (isCancelled || !mapInstance || !markersLayerRef.current) return;

      const map = mapInstance;
      const markersLayer = markersLayerRef.current;
      markersLayer.clearLayers();

      // Barcha tasdiqlangan klinikalarni koordinatalar bilan ta'minlash (agar aniq koordinata bo'lmasa, URL yoki viloyat koordinatasi)
      const validClinicsWithCoords = clinics
        .map((c) => {
          let lat = typeof c.latitude === 'number' && !isNaN(c.latitude) ? c.latitude : null;
          let lon = typeof c.longitude === 'number' && !isNaN(c.longitude) ? c.longitude : null;

          if (lat === null || lon === null) {
            const parsed = extractCoordinatesFromUrl((c as any).location_url);
            if (parsed) {
              lat = parsed.latitude;
              lon = parsed.longitude;
            } else if (c.region && REGION_DEFAULT_COORDINATES[c.region]) {
              lat = REGION_DEFAULT_COORDINATES[c.region].latitude;
              lon = REGION_DEFAULT_COORDINATES[c.region].longitude;
            }
          }

          return {
            ...c,
            latitude: lat,
            longitude: lon,
          };
        })
        .filter(
          (c): c is typeof c & { latitude: number; longitude: number } =>
            typeof c.latitude === 'number' && typeof c.longitude === 'number' && !isNaN(c.latitude) && !isNaN(c.longitude)
        );

      const bounds = L.latLngBounds([]);

      // 1. Foydalanuvchi joylashuvi marker (ko'k pulsatsiya)
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

      // 2. Picker rejimi (Klinika administratori tanlagan nuqta)
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
      validClinicsWithCoords.forEach((clinic) => {
        const isSelected = selectedClinicId === clinic.id;
        const distanceStr =
          'distance_km' in clinic && typeof (clinic as any).distance_km === 'number'
            ? `<div class="inline-flex items-center gap-1 text-[11px] font-semibold text-brand-teal bg-teal-50 px-2 py-0.5 rounded-md border border-teal-100 my-1">
                 <span>📍 ${(clinic as any).distance_km.toFixed(1)} km masofada</span>
               </div>`
            : '';

        const redPinSvg = isSelected
          ? `<svg width="38" height="48" viewBox="0 0 34 44" fill="none" xmlns="http://www.w3.org/2000/svg" style="filter: drop-shadow(0 4px 6px rgba(0,0,0,0.45));">
               <path d="M17 0C7.611 0 0 7.611 0 17C0 29.75 15.3 43.15 15.95 43.71C16.25 43.97 16.62 44.1 17 44.1C17.38 44.1 17.75 43.97 18.05 43.71C18.7 43.15 34 29.75 34 17C34 7.611 26.389 0 17 0Z" fill="#DC2626" stroke="#FFFFFF" stroke-width="2.5"/>
               <circle cx="17" cy="16" r="6.5" fill="#FFFFFF"/>
             </svg>`
          : `<svg width="32" height="42" viewBox="0 0 34 44" fill="none" xmlns="http://www.w3.org/2000/svg" style="filter: drop-shadow(0 3px 4px rgba(0,0,0,0.35));">
               <path d="M17 0C7.611 0 0 7.611 0 17C0 29.75 15.3 43.15 15.95 43.71C16.25 43.97 16.62 44.1 17 44.1C17.38 44.1 17.75 43.97 18.05 43.71C18.7 43.15 34 29.75 34 17C34 7.611 26.389 0 17 0Z" fill="#EA4335" stroke="#B91C1C" stroke-width="1.2"/>
               <circle cx="17" cy="16" r="5.5" fill="#FFFFFF"/>
             </svg>`;

        const clinicIcon = L.divIcon({
          className: 'qabulgo-red-pin',
          html: `
            <div style="position: relative; display: flex; flex-direction: column; align-items: center; justify-content: center; cursor: pointer; transform: ${
              isSelected ? 'scale(1.2)' : 'scale(1)'
            }; transition: transform 0.2s;">
              ${redPinSvg}
              <div style="width: 14px; height: 5px; background: rgba(0,0,0,0.25); border-radius: 50%; filter: blur(1px); margin-top: -2px;"></div>
              ${
                isSelected
                  ? `<div style="position: absolute; top: -28px; background: #08234A; color: white; font-size: 11px; font-weight: bold; padding: 2px 8px; border-radius: 6px; white-space: nowrap; box-shadow: 0 2px 4px rgba(0,0,0,0.25); pointer-events: none;">${clinic.name}</div>`
                  : ''
              }
            </div>
          `,
          iconSize: isSelected ? [38, 48] : [32, 42],
          iconAnchor: isSelected ? [19, 48] : [16, 42],
          popupAnchor: isSelected ? [0, -48] : [0, -42],
        });

        const marker = L.marker([clinic.latitude, clinic.longitude], {
          icon: clinicIcon,
          title: clinic.name,
        });

        const popupContent = `
          <div class="p-2 font-sans max-w-[240px]">
            <h4 class="font-bold text-sm text-brand-navy leading-snug line-clamp-1">${clinic.name}</h4>
            <p class="text-xs text-slate-500 mt-1 line-clamp-2">${clinic.address || clinic.district || clinic.region}</p>
            ${distanceStr}
            <div class="mt-2.5 pt-2 border-t border-slate-100 flex items-center justify-between gap-2">
              <button type="button" class="inline-flex items-center justify-center px-3 py-1 text-xs font-semibold rounded-lg bg-brand-blue text-white hover:bg-brand-blue-600 transition-colors">
                Tanlash
              </button>
              ${
                clinic.phone
                  ? `<span class="text-xs font-medium text-slate-600">
                       ${clinic.phone}
                     </span>`
                  : ''
              }
            </div>
          </div>
        `;

        marker.bindPopup(popupContent);

        marker.on('click', () => {
          if (onSelectClinic) {
            onSelectClinic(clinic.id);
          }
          map.panTo([clinic.latitude, clinic.longitude], { animate: true, duration: 0.5 });
        });

        markersLayer.addLayer(marker);
        bounds.extend([clinic.latitude, clinic.longitude]);

        if (isSelected) {
          marker.openPopup();
        }
      });

      // Fit bounds if we have valid coordinates and not single-point zoom
      if (bounds.isValid() && validClinicsWithCoords.length > 1) {
        map.fitBounds(bounds, { padding: [50, 50], maxZoom: 15 });
      } else if (validClinicsWithCoords.length === 1) {
        map.setView([validClinicsWithCoords[0].latitude, validClinicsWithCoords[0].longitude], 14);
      }
    }

    renderLayers();

    return () => {
      isCancelled = true;
    };
  }, [mapInstance, clinics, userLocation, pickedLocation, selectedClinicId, pickerMode]);

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

