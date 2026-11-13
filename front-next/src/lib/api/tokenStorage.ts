/**
 * Token storage with 30-day persistence for Patient, Clinic, and Doctor roles,
 * and strict session-only (never persistent) storage for Super Admin.
 */

type TokenListener = (token: string | null) => void;

let inMemoryToken: string | null = null;
const listeners: Set<TokenListener> = new Set();

const TOKEN_KEY = 'qabulgo_access_token';
const TOKEN_TIME_KEY = 'qabulgo_token_time';
const ADMIN_SESSION_TOKEN_KEY = 'qabulgo_admin_access_token';
const MAX_SESSION_MS = 30 * 24 * 60 * 60 * 1000; // 30 days in milliseconds

function parseJwtRole(token: string): string | null {
  try {
    const base64Url = token.split('.')[1];
    if (!base64Url) return null;
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );
    const parsed = JSON.parse(jsonPayload);
    return parsed.role || null;
  } catch {
    return null;
  }
}

export const tokenStorage = {
  getAccessToken(): string | null {
    if (inMemoryToken) {
      return inMemoryToken;
    }

    if (typeof window === 'undefined') {
      return null;
    }

    // 1. Check session storage for super admin (session-only, cleared on browser close)
    try {
      const sessionToken = sessionStorage.getItem(ADMIN_SESSION_TOKEN_KEY);
      if (sessionToken) {
        inMemoryToken = sessionToken;
        return sessionToken;
      }
    } catch {
      // Ignore
    }

    // 2. Check local storage for persistent 30-day users (patient, doctor, clinic)
    try {
      const localToken = localStorage.getItem(TOKEN_KEY);
      const savedTime = localStorage.getItem(TOKEN_TIME_KEY);

      if (localToken && savedTime) {
        const age = Date.now() - parseInt(savedTime, 10);
        if (age < MAX_SESSION_MS) {
          // Double check it's not a super_admin token mistakenly in localStorage
          const role = parseJwtRole(localToken);
          if (role === 'super_admin') {
            // Strictly forbid super_admin in localStorage
            localStorage.removeItem(TOKEN_KEY);
            localStorage.removeItem(TOKEN_TIME_KEY);
            return null;
          }
          inMemoryToken = localToken;
          return localToken;
        } else {
          // 30 days expired
          localStorage.removeItem(TOKEN_KEY);
          localStorage.removeItem(TOKEN_TIME_KEY);
        }
      }
    } catch {
      // Ignore
    }

    return null;
  },

  setAccessToken(token: string | null, explicitRole?: string): void {
    inMemoryToken = token;

    if (typeof window !== 'undefined') {
      try {
        if (!token) {
          localStorage.removeItem(TOKEN_KEY);
          localStorage.removeItem(TOKEN_TIME_KEY);
          sessionStorage.removeItem(ADMIN_SESSION_TOKEN_KEY);
        } else {
          const role = explicitRole || parseJwtRole(token);
          if (role === 'super_admin') {
            // Super Admin: NEVER save to localStorage! Store only in sessionStorage for active tab/session
            sessionStorage.setItem(ADMIN_SESSION_TOKEN_KEY, token);
            localStorage.removeItem(TOKEN_KEY);
            localStorage.removeItem(TOKEN_TIME_KEY);
          } else {
            // Patient, Doctor, Clinic Owner/Admin: Persist for 30 days in localStorage
            localStorage.setItem(TOKEN_KEY, token);
            if (!localStorage.getItem(TOKEN_TIME_KEY)) {
              localStorage.setItem(TOKEN_TIME_KEY, Date.now().toString());
            }
            sessionStorage.removeItem(ADMIN_SESSION_TOKEN_KEY);
          }
        }
      } catch (err) {
        console.error('Storage access error:', err);
      }
    }

    listeners.forEach((listener) => {
      try {
        listener(token);
      } catch (err) {
        console.error('Error notifying token listener', err);
      }
    });
  },

  clear(): void {
    this.setAccessToken(null);
    if (typeof window !== 'undefined') {
      try {
        localStorage.removeItem(TOKEN_KEY);
        localStorage.removeItem(TOKEN_TIME_KEY);
        localStorage.removeItem('qabulgo_user');
        sessionStorage.removeItem(ADMIN_SESSION_TOKEN_KEY);
        sessionStorage.removeItem('qabulgo_user');
      } catch {
        // Ignore
      }
    }
  },

  hasToken(): boolean {
    return Boolean(this.getAccessToken());
  },

  subscribe(listener: TokenListener): () => void {
    listeners.add(listener);
    return () => {
      listeners.delete(listener);
    };
  },
};
