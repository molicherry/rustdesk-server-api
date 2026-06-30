import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

import en from './locales/en'
import zhCN from './locales/zh-CN'

void i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: { en, 'zh-CN': zhCN },
    fallbackLng: 'en',
    supportedLngs: ['en', 'zh-CN'],
    defaultNS: 'common',
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
    },
    interpolation: {
      escapeValue: false,
    },
  })
  .then(() => {
    document.documentElement.lang = i18n.language
  })

export default i18n
