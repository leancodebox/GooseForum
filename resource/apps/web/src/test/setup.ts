import { supportedLocales } from "@gooseforum/client/i18n/locale";
import {
  gooseNamespaces,
  prepareGooseTranslations,
} from "@gooseforum/runtime/i18n";
import { prepareAdminTranslations } from "../admin/translation-loader";
import { adminNamespaces } from "@gooseforum/i18n/admin";

// Unit fixtures render pages directly; real entry points prepare only their active route.
await Promise.all(
  supportedLocales.flatMap((locale) => [
    prepareGooseTranslations(locale, gooseNamespaces),
    prepareAdminTranslations(locale, adminNamespaces),
  ]),
);
