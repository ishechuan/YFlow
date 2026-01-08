import {
  defineConfig,
  presetAttributify,
  presetIcons,
  presetUno,
  presetWebFonts,
  transformerDirectives,
  transformerVariantGroup,
} from 'unocss'

export default defineConfig({
  presets: [
    presetUno(),
    presetAttributify(),
    presetIcons({
      scale: 1.2,
      warn: true,
      extraProperties: {
        'display': 'inline-block',
        'vertical-align': 'middle',
      },
    }),
    presetWebFonts({
      fonts: {
        sans: 'Inter:-300..700',
        mono: 'JetBrains Mono:-300..700',
      },
    }),
  ],
  transformers: [
    transformerDirectives(),
    transformerVariantGroup(),
  ],
  theme: {
    colors: {
      // Deep Aurora Theme - 深海极光配色
      primary: {
        DEFAULT: '#0891b2',
        50: '#ecfeff',
        100: '#cffafe',
        200: '#a5f3fc',
        300: '#67e8f9',
        400: '#22d3ee',
        500: '#06b6d4',
        600: '#0891b2',
        700: '#0e7490',
        800: '#155e75',
        900: '#164e63',
        950: '#083344',
      },
      accent: {
        DEFAULT: '#14b8a6',
        50: '#f0fdfa',
        100: '#ccfbf1',
        200: '#99f6e4',
        300: '#5eead4',
        400: '#2dd4bf',
        500: '#14b8a6',
        600: '#0d9488',
        700: '#0f766e',
        800: '#115e59',
        900: '#134e4a',
      },
      midnight: {
        50: '#f8fafc',
        100: '#f1f5f9',
        200: '#e2e8f0',
        300: '#cbd5e1',
        400: '#94a3b8',
        500: '#64748b',
        600: '#475569',
        700: '#334155',
        800: '#1e293b',
        900: '#0f172a',
        950: '#020617',
      },
      ocean: {
        DEFAULT: '#0c4a6e',
        light: '#075985',
        DEFAULT: '#0c4a6e',
        dark: '#082f49',
        deeper: '#020617',
      },
    },
    fontFamily: {
      sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
      mono: ['JetBrains Mono', 'Monaco', 'Consolas', 'monospace'],
    },
  },
  shortcuts: {
    'flex-center': 'flex items-center justify-center',
    'flex-between': 'flex items-center justify-between',
    'card-base': 'bg-white rounded-2xl shadow-sm border border-slate-100',
    'card-hover': 'hover:shadow-xl hover:-translate-y-1 transition-all duration-300',
    'btn-base': 'px-5 py-2.5 rounded-xl font-medium transition-all duration-200',
    'btn-primary': 'bg-gradient-to-r from-cyan-500 to-teal-500 hover:from-cyan-600 hover:to-teal-600 text-white btn-base shadow-lg shadow-cyan-500/30',
    'btn-secondary': 'bg-slate-100 hover:bg-slate-200 text-slate-700 btn-base',
    'input-base': 'w-full px-4 py-3 rounded-xl border border-slate-200 focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 transition-all',
  },
  rules: [
    ['text-gradient-aurora', {
      'background': 'linear-gradient(135deg, #06b6d4 0%, #14b8a6 50%, #22d3ee 100%)',
      '-webkit-background-clip': 'text',
      '-webkit-text-fill-color': 'transparent',
    }],
    ['text-gradient-ocean', {
      'background': 'linear-gradient(135deg, #0c4a6e 0%, #075985 100%)',
      '-webkit-background-clip': 'text',
      '-webkit-text-fill-color': 'transparent',
    }],
    ['bg-gradient-aurora', {
      'background': 'linear-gradient(135deg, #06b6d4 0%, #14b8a6 100%)',
    }],
    ['bg-gradient-midnight', {
      'background': 'linear-gradient(180deg, #0f172a 0%, #020617 100%)',
    }],
    ['bg-gradient-ocean', {
      'background': 'linear-gradient(135deg, #0c4a6e 0%, #082f49 50%, #020617 100%)',
    }],
    ['glass-effect', {
      'background': 'rgba(255, 255, 255, 0.7)',
      'backdrop-filter': 'blur(20px)',
      '-webkit-backdrop-filter': 'blur(20px)',
      'border': '1px solid rgba(255, 255, 255, 0.3)',
    }],
    ['glass-dark', {
      'background': 'rgba(15, 23, 42, 0.8)',
      'backdrop-filter': 'blur(20px)',
      '-webkit-backdrop-filter': 'blur(20px)',
      'border': '1px solid rgba(255, 255, 255, 0.1)',
    }],
    ['glow-aurora', {
      'box-shadow': '0 0 40px rgba(6, 182, 212, 0.3), 0 0 80px rgba(20, 184, 166, 0.2)',
    }],
    ['glow-text', {
      'text-shadow': '0 0 20px rgba(6, 182, 212, 0.5)',
    }],
  ],
})
