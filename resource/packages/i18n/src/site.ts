import { loaders, namespaces } from '../generated/site'
import { loaders as sharedLoaders, namespaces as sharedNamespaces } from '../generated/shared'
import { createTranslationStore } from './index'

export const siteLoaders = { ...loaders, ...sharedLoaders }
export const siteNamespaces = [...namespaces, ...sharedNamespaces]
export type SiteNamespace = (typeof siteNamespaces)[number]
export const siteTranslations = createTranslationStore<SiteNamespace>(siteLoaders)
