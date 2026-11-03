import type { Metadata, Viewport } from 'next';
import '@/styles/globals.css';
import { AppProviders } from '@/providers';

export const metadata: Metadata = {
  metadataBase: new URL('https://qabulgo.uz'),
  title: {
    default: 'QabulGo — Shifokor va klinikalarga onlayn qabul',
    template: '%s | QabulGo',
  },
  description:
    'Termiz va O‘zbekiston bo‘yicha xususiy klinikalar, malakali shifokorlar (nevropatolog, urolog, kardiolog va b.), qabul narxlari va onlayn navbatga yozilish platformasi.',
  applicationName: 'QabulGo',
  keywords: [
    'QabulGo',
    'Termiz klinikalari',
    'Termizdagi xususiy klinikalar',
    'Termiz nevropatolog',
    'Termiz urolog',
    'Termiz kardiolog',
    'Termiz stomatolog',
    'Termiz pediatr',
    'Shifokor qabuliga yozilish',
    'O‘zbekistondagi xususiy klinikalar',
    'Shifokorlarni onlayn bron qilish',
    'Onlayn navbat shifokor',
    'Surxondaryo shifoxonalari',
    'Tibbiy markazlar Termiz',
  ],
  authors: [{ name: 'QabulGo Jamoasi', url: 'https://qabulgo.uz' }],
  creator: 'QabulGo',
  publisher: 'QabulGo',
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  alternates: {
    canonical: '/',
  },
  openGraph: {
    title: 'QabulGo — Shifokor va klinikalarga onlayn qabul',
    description:
      'Termiz va O‘zbekiston xususiy klinikalari, shifokorlar, narxlar va navbatsiz onlayn qabul.',
    url: 'https://qabulgo.uz',
    siteName: 'QabulGo',
    locale: 'uz_UZ',
    type: 'website',
    images: [
      {
        url: '/logo-mark.png',
        width: 512,
        height: 512,
        alt: 'QabulGo — Tibbiyot platformasi',
      },
    ],
  },
  twitter: {
    card: 'summary',
    title: 'QabulGo — Shifokor va klinikalarga onlayn qabul',
    description:
      'Termiz va O‘zbekiston xususiy klinikalari, shifokorlar, narxlar va navbatsiz onlayn qabul.',
    images: ['/logo-mark.png'],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  icons: {
    icon: '/logo.svg',
    shortcut: '/logo-mark.png',
    apple: '/logo-mark.png',
  },
};

export const viewport: Viewport = {
  themeColor: '#0866F5',
  width: 'device-width',
  initialScale: 1,
  maximumScale: 5,
};

const orgJsonLd = {
  '@context': 'https://schema.org',
  '@type': 'MedicalBusiness',
  name: 'QabulGo',
  url: 'https://qabulgo.uz',
  logo: 'https://qabulgo.uz/logo-mark.png',
  description:
    'O‘zbekistondagi xususiy klinikalar va shifokorlarni qidirish, narxlarni ko‘rish va onlayn navbatga yozilish platformasi.',
  address: {
    '@type': 'PostalAddress',
    addressLocality: 'Termiz',
    addressRegion: 'Surxondaryo viloyati',
    addressCountry: 'UZ',
  },
  areaServed: {
    '@type': 'Country',
    name: 'Uzbekistan',
  },
};

const websiteJsonLd = {
  '@context': 'https://schema.org',
  '@type': 'WebSite',
  name: 'QabulGo',
  url: 'https://qabulgo.uz',
  potentialAction: {
    '@type': 'SearchAction',
    target: {
      '@type': 'EntryPoint',
      urlTemplate: 'https://qabulgo.uz/clinics?search={search_term_string}',
    },
    'query-input': 'required name=search_term_string',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="uz" className="h-full bg-brand-bg font-sans">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap"
          rel="stylesheet"
        />
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(orgJsonLd) }}
        />
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(websiteJsonLd) }}
        />
      </head>
      <body className="h-full antialiased text-brand-navy selection:bg-brand-blue-100 selection:text-brand-blue-900">
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
