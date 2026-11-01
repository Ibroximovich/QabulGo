import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import axios from 'axios';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      gcTime: 1000 * 60 * 30, // 30 minutes
      refetchOnWindowFocus: false,
      retry: (failureCount, error) => {
        if (failureCount >= 2) return false;
        if (axios.isAxiosError(error)) {
          const status = error.response?.status;
          // Don't retry client errors
          if (status && [400, 401, 403, 404, 409, 422].includes(status)) {
            return false;
          }
        }
        return true;
      },
    },
    mutations: {
      retry: false,
    },
  },
});

export function QueryProvider({ children }: { children: React.ReactNode }) {
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

