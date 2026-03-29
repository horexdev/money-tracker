import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

import en from './locales/en.json'
import ru from './locales/ru.json'
import uk from './locales/uk.json'
import be from './locales/be.json'
import kk from './locales/kk.json'
import uz from './locales/uz.json'
import es from './locales/es.json'
import de from './locales/de.json'
import it from './locales/it.json'
import fr from './locales/fr.json'
import pt from './locales/pt.json'
import nl from './locales/nl.json'
import ar from './locales/ar.json'
import tr from './locales/tr.json'
import ko from './locales/ko.json'
import ms from './locales/ms.json'
import id from './locales/id.json'

const SUPPORTED_LANGS = ['en', 'ru', 'uk', 'be', 'kk', 'uz', 'es', 'de', 'it', 'fr', 'pt', 'nl', 'ar', 'tr', 'ko', 'ms', 'id']

/** Try to extract the user's language from Telegram WebApp initData */
function getTelegramLanguage(): string | undefined {
  try {
    // Telegram Mini Apps pass initData as a URL search string in the hash
    const hash = window.location.hash
    const initDataRaw = new URLSearchParams(hash.includes('tgWebAppData=')
      ? hash.slice(hash.indexOf('tgWebAppData='))
      : ''
    ).get('tgWebAppData')

    if (initDataRaw) {
      const params = new URLSearchParams(initDataRaw)
      const userJson = params.get('user')
      if (userJson) {
        const user = JSON.parse(userJson)
        if (user.language_code && SUPPORTED_LANGS.includes(user.language_code)) {
          return user.language_code
        }
      }
    }

    // Fallback: try window.Telegram.WebApp
    const tg = (window as unknown as { Telegram?: { WebApp?: { initDataUnsafe?: { user?: { language_code?: string } } } } }).Telegram
    const langCode = tg?.WebApp?.initDataUnsafe?.user?.language_code
    if (langCode && SUPPORTED_LANGS.includes(langCode)) {
      return langCode
    }
  } catch {
    // ignore
  }
  return undefined
}

// Custom language detector for Telegram
const telegramDetector = {
  name: 'telegramDetector',
  lookup(): string | undefined {
    return getTelegramLanguage()
  },
}

const languageDetector = new LanguageDetector()
languageDetector.addDetector(telegramDetector)

i18n
  .use(languageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: en },
      ru: { translation: ru },
      uk: { translation: uk },
      be: { translation: be },
      kk: { translation: kk },
      uz: { translation: uz },
      es: { translation: es },
      de: { translation: de },
      it: { translation: it },
      fr: { translation: fr },
      pt: { translation: pt },
      nl: { translation: nl },
      ar: { translation: ar },
      tr: { translation: tr },
      ko: { translation: ko },
      ms: { translation: ms },
      id: { translation: id },
    },
    fallbackLng: 'en',
    supportedLngs: SUPPORTED_LANGS,
    interpolation: {
      escapeValue: false,
    },
    detection: {
      order: ['querystring', 'localStorage', 'telegramDetector', 'navigator'],
      lookupQuerystring: 'lang',
      lookupLocalStorage: 'i18nextLng',
      caches: ['localStorage'],
    },
  })

export default i18n
