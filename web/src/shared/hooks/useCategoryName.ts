import { useTranslation } from 'react-i18next'

/** Returns a function that translates a category name if a translation exists, otherwise returns the original name. */
export function useCategoryName() {
  const { t } = useTranslation()
  return (name: string) => t(`categories.names.${name}`, { defaultValue: name })
}
