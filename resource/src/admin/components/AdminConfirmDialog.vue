<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'
import { AlertTriangle } from '@lucide/vue'
import { Button } from '@/admin/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/admin/components/ui/dialog'

withDefaults(defineProps<{
  cancelText?: string
  confirmText?: string
  description?: string
  loading?: boolean
  open: boolean
  title: string
}>(), {
  cancelText: adminText('k009q'),
  confirmText: adminText('k005i'),
  description: '',
  loading: false,
})

const emit = defineEmits<{
  confirm: []
  'update:open': [value: boolean]
}>()
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <div class="flex items-start gap-3">
          <span class="mt-0.5 grid size-9 shrink-0 place-items-center rounded-md bg-destructive/10 text-destructive">
            <AlertTriangle class="size-4" />
          </span>
          <span class="min-w-0">
            <DialogTitle>{{ title }}</DialogTitle>
            <DialogDescription v-if="description" class="mt-1">
              {{ description }}
            </DialogDescription>
          </span>
        </div>
      </DialogHeader>
      <DialogFooter>
        <Button variant="outline" type="button" @click="emit('update:open', false)">{{ cancelText }}</Button>
        <Button variant="destructive" type="button" :disabled="loading" @click="emit('confirm')">
          {{ loading ? adminText('k005h') : confirmText }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
