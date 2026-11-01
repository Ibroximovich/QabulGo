import type { Metadata, Viewport } from 'next';
import '@/styles/globals.css';
import { AppProviders } from '@/providers';

export const metadata: Metadata = {
  title: 'QabulGo — Shifokor va klinikalarga onlayn qabul',
  description:
    'QabulGo — O‘zbekistondagi xususiy klinikalar va shifokorlarga onlayn qabulga yozilish platformasi.',
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
      </head>
      <body className="h-full antialiased text-brand-navy selection:bg-brand-blue-100 selection:text-brand-blue-900">
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
