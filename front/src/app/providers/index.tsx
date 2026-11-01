import React from 'react';
import { QueryProvider } from './QueryProvider';
import { AuthProvider } from './AuthProvider';
import { Toaster } from 'sonner';

export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <QueryProvider>
      <AuthProvider>
        {children}
        <Toaster
          position="top-right"
          richColors
          closeButton
          duration={4000}
          toastOptions={{
            style: {
              borderRadius: '1rem',
              fontSize: '0.875rem',
            },
          }}
        />
      </AuthProvider>
    </QueryProvider>
  );
}

