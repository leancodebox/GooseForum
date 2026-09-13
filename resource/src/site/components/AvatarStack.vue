<script setup lang="ts">
import UserAvatar from '@/site/components/UserAvatar.vue'
import UserCardPopover from '@/site/components/UserCardPopover.vue'

interface StackUser {
  id: number
  username: string
  avatarUrl: string
}

withDefaults(defineProps<{
  users: StackUser[]
  size?: 'sm' | 'md'
}>(), {
  size: 'md',
})
</script>

<template>
  <div
    class="flex"
    :class="size === 'sm' ? 'h-6 min-w-6 -space-x-2' : 'h-8 min-w-8 -space-x-3'"
  >
    <UserCardPopover
      v-for="user in users"
      :key="user.id"
      :user="user"
    >
      <a
        :href="`/u/${user.id}`"
        :title="user.username"
        class="rounded-full ring-2 ring-base-100 transition hover:z-10 hover:scale-110"
        :class="size === 'sm' ? 'h-6 w-6' : 'h-8 w-8'"
      >
        <UserAvatar
          :src="user.avatarUrl"
          :alt="user.username"
          class="rounded-full object-cover"
          :class="size === 'sm' ? 'h-6 w-6' : 'h-8 w-8'"
        />
      </a>
    </UserCardPopover>
  </div>
</template>
