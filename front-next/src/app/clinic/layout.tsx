import type { Metadata } from 'next';

export const metadata: Metadata = {
  robots: {
    index: false,
    follow: false,
    nocache: true,
  },
};

export default function ClinicRootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}

