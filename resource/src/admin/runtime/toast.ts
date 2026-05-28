import { adminText } from '@/admin/runtime/i18n-text'
import { toast } from 'vue-sonner'

function message(error: unknown, fallback: string) {
  return error instanceof Error && error.message ? error.message : fallback
}

export const adminToast = {
  success(text: string) {
    toast.success(text)
  },
  error(error: unknown, fallback = adminText('k001u')) {
    toast.error(message(error, fallback))
  },
  info(text: string) {
    toast.info(text)
  },
  warning(text: string) {
    toast.warning(text)
  },
}
