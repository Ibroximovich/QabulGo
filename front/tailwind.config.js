/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          blue: {
            DEFAULT: '#0866F5',
            50: '#EEF5FE',
            100: '#D9E8FD',
            200: '#B4D2FB',
            300: '#7FADF9',
            400: '#4384F7',
            500: '#0866F5',
            600: '#0652C7',
            700: '#05409C',
            800: '#042E6F',
            900: '#031E48',
          },
          teal: {
            DEFAULT: '#08BDBA',
            50: '#E6FAF9',
            100: '#C2F4F2',
            200: '#8AE8E5',
            300: '#52DDD8',
            400: '#1AD1CB',
            500: '#08BDBA',
            600: '#069B98',
            700: '#057A77',
            800: '#045957',
            900: '#023837',
          },
          navy: {
            DEFAULT: '#08234A',
            50: '#F0F4F9',
            100: '#DBE4F0',
            200: '#B6C9E1',
            300: '#86A7CE',
            400: '#5380B7',
            500: '#32609B',
            600: '#23497C',
            700: '#19365E',
            800: '#0F2543',
            900: '#08234A',
          },
          bg: '#F7FAFC',
          muted: '#64748B',
          success: '#16A34A',
          error: '#DC2626',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
      boxShadow: {
        subtle: '0 1px 3px 0 rgba(8, 35, 74, 0.05), 0 1px 2px -1px rgba(8, 35, 74, 0.05)',
        card: '0 4px 6px -1px rgba(8, 35, 74, 0.07), 0 2px 4px -2px rgba(8, 35, 74, 0.05)',
        cardHover: '0 10px 15px -3px rgba(8, 35, 74, 0.1), 0 4px 6px -4px rgba(8, 35, 74, 0.08)',
        elevated: '0 20px 25px -5px rgba(8, 35, 74, 0.1), 0 8px 10px -6px rgba(8, 35, 74, 0.08)',
        modal: '0 25px 50px -12px rgba(8, 35, 74, 0.25)',
      },
      borderRadius: {
        'xl': '0.75rem',
        '2xl': '1rem',
        '3xl': '1.5rem',
      },
      animation: {
        'fade-in': 'fadeIn 0.2s ease-in-out',
        'slide-up': 'slideUp 0.3s cubic-bezier(0.16, 1, 0.3, 1)',
        'pulse-subtle': 'pulseSubtle 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(8px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        pulseSubtle: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.85' },
        },
      },
    },
  },
  plugins: [],
}

