/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Outfit', 'system-ui', 'sans-serif'],
      },
      colors: {
        cinema: {
          dark: '#0c0a09',
          card: '#1c1917',
          border: '#292524',
          muted: '#78716c',
          gold: '#f59e0b',
          screen: '#22c55e',
        },
      },
    },
  },
  plugins: [],
}
