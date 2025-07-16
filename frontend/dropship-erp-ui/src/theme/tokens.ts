// Design tokens for consistent styling across the application
export const colors = {
  primary: '#1976d2',
  secondary: '#dc004e',
  background: '#ffffff',
  surface: '#f5f5f5',
  text: '#1f2937',
  textSecondary: '#6b7280',
  error: '#ef4444',
  warning: '#f59e0b',
  success: '#10b981',
  info: '#3b82f6',
  divider: '#e5e7eb',
} as const;

export const spacing = {
  xs: '0.25rem',  // 4px
  sm: '0.5rem',   // 8px
  md: '1rem',     // 16px
  lg: '1.5rem',   // 24px
  xl: '2rem',     // 32px
  xxl: '3rem',    // 48px
} as const;

export const typography = {
  fontFamily: '"Roboto", "Helvetica", "Arial", sans-serif',
  fontSize: {
    xs: '0.75rem',    // 12px
    sm: '0.875rem',   // 14px
    base: '1rem',     // 16px
    lg: '1.125rem',   // 18px
    xl: '1.25rem',    // 20px
    xxl: '1.5rem',    // 24px
  },
  fontWeight: {
    normal: 400,
    medium: 500,
    semibold: 600,
    bold: 700,
  },
} as const;

export const shadows = {
  none: 'none',
  sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
  base: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
  md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
  lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
} as const;

export const borderRadius = {
  none: '0',
  sm: '0.125rem',   // 2px
  base: '0.25rem',  // 4px
  md: '0.375rem',   // 6px
  lg: '0.5rem',     // 8px
  xl: '0.75rem',    // 12px
  full: '9999px',
} as const;

export const breakpoints = {
  xs: '0px',
  sm: '600px',
  md: '900px',
  lg: '1200px',
  xl: '1536px',
} as const;

// CSS custom properties for dynamic theming
export const cssVariables = {
  '--color-primary': colors.primary,
  '--color-secondary': colors.secondary,
  '--color-background': colors.background,
  '--color-surface': colors.surface,
  '--color-text': colors.text,
  '--color-text-secondary': colors.textSecondary,
  '--color-error': colors.error,
  '--color-warning': colors.warning,
  '--color-success': colors.success,
  '--color-info': colors.info,
  '--color-divider': colors.divider,
  '--spacing-xs': spacing.xs,
  '--spacing-sm': spacing.sm,
  '--spacing-md': spacing.md,
  '--spacing-lg': spacing.lg,
  '--spacing-xl': spacing.xl,
  '--spacing-xxl': spacing.xxl,
  '--font-family': typography.fontFamily,
  '--border-radius-base': borderRadius.base,
  '--shadow-base': shadows.base,
} as const;