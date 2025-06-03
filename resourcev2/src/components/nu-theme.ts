
import { NConfigProvider, GlobalThemeOverrides, darkTheme, lightTheme } from 'naive-ui'

// DaisyUI主题颜色映射
const themeColorMaps: Record<string, Record<string, string>> = {
  light: {
    primary: '#8b5cf6', // oklch(45% 0.24 277.023)
    primaryContent: '#f3f4f6', // oklch(93% 0.034 272.788)
    base100: '#ffffff', // oklch(100% 0 0)
    base200: '#f9fafb', // oklch(98% 0 0)
    base300: '#f3f4f6', // oklch(95% 0 0)
    baseContent: '#374151' // oklch(21% 0.006 285.885)
  },
  dark: {
    primary: '#1FB2A5', // oklch(58% 0.233 277.117)
    primaryContent: '#F0F9FF', // oklch(96% 0.018 272.314)
    base100: '#1D232A', // oklch(25.33% 0.016 252.42)
    base200: '#191E24', // oklch(23.26% 0.014 253.1)
    base300: '#15191E', // oklch(21.15% 0.012 254.09)
    baseContent: '#A6ADBB' // oklch(97.807% 0.029 256.847)
  },
  cupcake: {
    primary: '#65C3C8', // oklch(85% 0.138 181.071)
    primaryContent: '#291334', // oklch(43% 0.078 188.216)
    base100: '#FAF7F5', // oklch(97.788% 0.004 56.375)
    base200: '#EFEAE6', // oklch(93.982% 0.007 61.449)
    base300: '#E7E2DF', // oklch(91.586% 0.006 53.44)
    baseContent: '#291334' // oklch(23.574% 0.066 313.189)
  },
  bumblebee: {
    primary: '#F9D72F', // oklch(85% 0.199 91.936)
    primaryContent: '#181830', // oklch(42% 0.095 57.708)
    base100: '#FFFFFF', // oklch(100% 0 0)
    base200: '#F7F7F7', // oklch(97% 0 0)
    base300: '#EBEBEB', // oklch(92% 0 0)
    baseContent: '#333333' // oklch(20% 0 0)
  },
  emerald: {
    primary: '#66CC8A', // oklch(76.662% 0.135 153.45)
    primaryContent: '#223344', // oklch(33.387% 0.04 162.24)
    base100: '#FFFFFF', // oklch(100% 0 0)
    base200: '#EDEDED', // oklch(93% 0 0)
    base300: '#DBDBDB', // oklch(86% 0 0)
    baseContent: '#394E6A' // oklch(35.519% 0.032 262.988)
  },
  corporate: {
    primary: '#4B6BFB', // oklch(58% 0.158 241.966)
    primaryContent: '#FFFFFF', // oklch(100% 0 0)
    base100: '#FFFFFF', // oklch(100% 0 0)
    base200: '#EDEDED', // oklch(93% 0 0)
    base300: '#DBDBDB', // oklch(86% 0 0)
    baseContent: '#181A2A' // oklch(22.389% 0.031 278.072)
  },
  synthwave: {
    primary: '#E779C1', // oklch(71% 0.202 349.761)
    primaryContent: '#19072C', // oklch(28% 0.109 3.907)
    base100: '#2D1B69', // oklch(15% 0.09 281.288)
    base200: '#3A2A7C', // oklch(20% 0.09 281.288)
    base300: '#4A3A8C', // oklch(25% 0.09 281.288)
    baseContent: '#A991F7' // oklch(78% 0.115 274.713)
  },
  retro: {
    primary: '#EF9995', // oklch(80% 0.114 19.571)
    primaryContent: '#282425', // oklch(39% 0.141 25.723)
    base100: '#ECE3CA', // oklch(91.637% 0.034 90.515)
    base200: '#E4D8B4', // oklch(88.272% 0.049 91.774)
    base300: '#DACC9E', // oklch(84.133% 0.065 90.856)
    baseContent: '#282425' // oklch(41% 0.112 45.904)
  },
  cyberpunk: {
    primary: '#FF7598', // oklch(74.22% 0.209 6.35)
    primaryContent: '#000000', // oklch(14.844% 0.041 6.35)
    base100: '#FFE300', // oklch(94.51% 0.179 104.32)
    base200: '#FFD700', // oklch(91.51% 0.179 104.32)
    base300: '#FFCC00', // oklch(85.51% 0.179 104.32)
    baseContent: '#000000' // oklch(0% 0 0)
  },
  valentine: {
    primary: '#E96D7B', // oklch(65% 0.241 354.308)
    primaryContent: '#FFFFFF', // oklch(100% 0 0)
    base100: '#F8DADA', // oklch(97% 0.014 343.198)
    base200: '#F2B6B6', // oklch(94% 0.028 342.258)
    base300: '#E96D7B', // oklch(89% 0.061 343.231)
    baseContent: '#632C3B' // oklch(52% 0.223 3.958)
  },
  halloween: {
    primary: '#F28C18', // oklch(77.48% 0.204 60.62)
    primaryContent: '#1A0A00', // oklch(19.693% 0.004 196.779)
    base100: '#212121', // oklch(21% 0.006 56.043)
    base200: '#1A1A1A', // oklch(14% 0.004 49.25)
    base300: '#000000', // oklch(0% 0 0)
    baseContent: '#D9D9D9' // oklch(84.955% 0 0)
  },
  garden: {
    primary: '#5C7F67', // oklch(62.45% 0.278 3.836)
    primaryContent: '#FFFFFF', // oklch(100% 0 0)
    base100: '#E8E8E8', // oklch(92.951% 0.002 17.197)
    base200: '#D1D1D1', // oklch(86.445% 0.002 17.197)
    base300: '#BABABA', // oklch(79.938% 0.001 17.197)
    baseContent: '#1F2937' // oklch(16.961% 0.001 17.32)
  },
  forest: {
    primary: '#1EB854', // oklch(68.628% 0.185 148.958)
    primaryContent: '#000000', // oklch(0% 0 0)
    base100: '#171212', // oklch(20.84% 0.008 17.911)
    base200: '#0D1117', // oklch(18.522% 0.007 17.911)
    base300: '#0A0E13', // oklch(16.203% 0.007 17.911)
    baseContent: '#C9D1D9' // oklch(83.768% 0.001 17.911)
  },
  aqua: {
    primary: '#09ECDA', // oklch(85.661% 0.144 198.645)
    primaryContent: '#003C3C', // oklch(40.124% 0.068 197.603)
    base100: '#345DA7', // oklch(37% 0.146 265.522)
    base200: '#1A2332', // oklch(28% 0.091 267.935)
    base300: '#0F1419', // oklch(22% 0.091 267.935)
    baseContent: '#B8E6FF' // oklch(90% 0.058 230.902)
  }
}

// 获取当前主题
const getCurrentTheme = (): string => {
  if (typeof window === 'undefined') return 'light'
  return document.documentElement.getAttribute('data-theme') || 'light'
}

// 获取主题颜色
const getThemeColors = (theme: string) => {
  return themeColorMaps[theme] || themeColorMaps.light
}

// 判断是否为深色主题
const isDarkTheme = (theme: string): boolean => {
  const darkThemes = ['dark', 'synthwave', 'halloween', 'forest', 'aqua']
  return darkThemes.includes(theme)
}

// 创建Select主题配置
export const createSelectThemeOverrides = (theme?: string): GlobalThemeOverrides => {
  const currentTheme = theme || getCurrentTheme()
  const colors = getThemeColors(currentTheme)
  const isLight = !isDarkTheme(currentTheme)
  
  return {
    Select: {
      peers: {
        InternalSelection: {
          textColor: colors.baseContent,
          placeholderColor: colors.baseContent + '80', // 50% opacity
          color: colors.base100,
          colorActive: colors.base200,
          border: `1px solid ${colors.base300}`,
          borderHover: `1px solid ${colors.primary}`,
          borderActive: `1px solid ${colors.primary}`,
          borderFocus: `1px solid ${colors.primary}`,
          boxShadowFocus: `0 0 0 2px ${colors.primary}40`, // 25% opacity
          loadingColor: colors.primary,
          borderRadius: '0.5rem',
          fontWeight: '400'
        },
        InternalSelectMenu: {
          color: colors.base100,
          optionColorActive: colors.primary,
          optionTextColor: colors.baseContent,
          optionTextColorActive: colors.primaryContent,
          groupHeaderTextColor: colors.baseContent,
          borderRadius: '0.5rem',
        }
      }
    }
  }
}

// 默认主题配置
export const themeOverrides: GlobalThemeOverrides = createSelectThemeOverrides()

// 获取基础主题
export const getBaseTheme = (theme?: string) => {
  const currentTheme = theme || getCurrentTheme()
  return isDarkTheme(currentTheme) ? darkTheme : lightTheme
}