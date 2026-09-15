import { supportedLocales } from "@gooseforum/client/i18n/locale";
import {
  gooseNamespaces,
  prepareGooseTranslations,
} from "@gooseforum/runtime/i18n";
import { prepareAdminTranslations } from "../admin/translation-loader";
import { translationLoaders } from "../admin/translation-resources";

// Unit fixtures render pages directly; real entry points prepare only their active route.
const adminNamespaces = [
  ...new Set(Object.keys(translationLoaders).map((key) => key.split("/")[1])),
];
await Promise.all(
  supportedLocales.flatMap((locale) => [
    prepareGooseTranslations(locale, gooseNamespaces),
    prepareAdminTranslations(locale, adminNamespaces),
  ]),
);
