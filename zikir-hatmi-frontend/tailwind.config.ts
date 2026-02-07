import type { Config } from 'tailwindcss'

export default {
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        sans: ['Sora', 'Inter', 'DM Sans', 'system-ui', 'sans-serif'],
      },
      backgroundImage: {
        aurora:
          'radial-gradient(circle at top, rgba(76,29,149,0.35), transparent 45%), radial-gradient(circle at 15% 25%, rgba(14,165,233,0.45), transparent 30%), radial-gradient(circle at 85% 10%, rgba(236,72,153,0.35), transparent 35%)',
      },
    },
  },
} satisfies Config
