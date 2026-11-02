import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ClinicCard } from '../components/ClinicCard';
import { DoctorCard } from '../components/DoctorCard';
import { EmptySearchState } from '../components/EmptySearchState';
import { PublicClinicItem, NearbyClinicItem, PublicDoctorItem } from '../types';

describe('Public Search UI Components', () => {
  const mockClinic: PublicClinicItem = {
    id: 'c111-2222',
    name: 'Shifo Nur Med Clinic',
    description: 'Zamonaviy ko‘p tarmoqli klinika',
    phone: '+998712001122',
    region: 'Toshkent shahri',
    district: 'Yunusobod tumani',
    address: 'Amir Temur ko‘chasi, 15',
    logo_url: '',
    cover_image_url: '',
    doctors_count: 8,
    specialties: [
      { id: 's1', name: 'Kardiologiya', slug: 'kardiologiya' },
      { id: 's2', name: 'Stomatologiya', slug: 'stomatologiya' },
    ],
  };

  const mockNearbyClinic: NearbyClinicItem = {
    ...mockClinic,
    id: 'c333-4444',
    name: 'Yaqin Tibbiyot Markazi',
    latitude: 41.3111,
    longitude: 69.2405,
    distance_km: 1.4,
  };

  const mockDoctor: PublicDoctorItem = {
    clinic_doctor_id: 'cd-1',
    doctor_id: 'd-1',
    first_name: 'Anvar',
    last_name: 'Qodirov',
    photo_url: '',
    bio: 'Tajribali kardiolog',
    experience_years: 12,
    consultation_price: 150000,
    appointment_duration_minutes: 30,
    specialty: { id: 's1', name: 'Kardiologiya', slug: 'kardiologiya' },
    clinic: {
      id: 'c111-2222',
      name: 'Shifo Nur Med Clinic',
      region: 'Toshkent shahri',
      district: 'Yunusobod tumani',
      address: 'Amir Temur ko‘chasi, 15',
      phone: '+998712001122',
      logo_url: '',
    },
  };

  it('renders ClinicCard with clinic name, address, specialties, and doctor count', () => {
    render(<ClinicCard clinic={mockClinic} />);

    expect(screen.getByText('Shifo Nur Med Clinic')).toBeInTheDocument();
    expect(screen.getByText(/Amir Temur ko‘chasi, 15/)).toBeInTheDocument();
    expect(screen.getByText('Kardiologiya')).toBeInTheDocument();
    expect(screen.getByText('Stomatologiya')).toBeInTheDocument();
    expect(screen.getByText('8 nafar shifokor')).toBeInTheDocument();
    expect(screen.getByText('Tasdiqlangan')).toBeInTheDocument();
  });

  it('renders ClinicCard with distance badge in nearby mode', () => {
    render(<ClinicCard clinic={mockNearbyClinic} />);

    expect(screen.getByText('Yaqin Tibbiyot Markazi')).toBeInTheDocument();
    expect(screen.getByText(/1.4 km yaqin/)).toBeInTheDocument();
  });

  it('renders DoctorCard with name, specialty, experience, formatted price, and clinic name', () => {
    render(<DoctorCard doctor={mockDoctor} />);

    expect(screen.getByText('Anvar Qodirov')).toBeInTheDocument();
    expect(screen.getByText('Kardiologiya')).toBeInTheDocument();
    expect(screen.getByText('12 yil tajriba')).toBeInTheDocument();
    expect(screen.getByText(/150[\s\u00A0]*000\s*so['‘]m/i)).toBeInTheDocument();
    expect(screen.getByText('Shifo Nur Med Clinic')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Qabulga yozilish/i })).toBeInTheDocument();
  });

  it('renders EmptySearchState with custom message and triggers onReset', () => {
    const handleReset = vi.fn();
    render(
      <EmptySearchState
        title="Klinika topilmadi"
        description="Boshqa mezonlarni tanlang"
        onReset={handleReset}
      />
    );

    expect(screen.getByText('Klinika topilmadi')).toBeInTheDocument();
    expect(screen.getByText('Boshqa mezonlarni tanlang')).toBeInTheDocument();

    const resetBtn = screen.getByRole('button', { name: /Filtrlarni tozalash/i });
    fireEvent.click(resetBtn);
    expect(handleReset).toHaveBeenCalledTimes(1);
  });
});
