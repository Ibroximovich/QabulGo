/**
 * In-memory token storage for maximum XSS security.
 * Access token is never written to localStorage or sessionStorage.
 * Refresh token is handled exclusively via HttpOnly cookies by the backend.
 */

type TokenListener = (token: string | null) => void;

let accessToken: string | null = null;
const listeners: Set<TokenListener> = new Set();

export const tokenStorage = {
  getAccessToken(): string | null {
    return accessToken;
  },

  setAccessToken(token: string | null): void {
    accessToken = token;
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
  },

  hasToken(): boolean {
    return Boolean(accessToken);
  },

  subscribe(listener: TokenListener): () => void {
    listeners.add(listener);
    return () => {
      listeners.delete(listener);
    };
  },
};

