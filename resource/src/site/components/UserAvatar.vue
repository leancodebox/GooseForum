<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  src: string
  alt: string
  size?: 'medium' | 'large'
}>(), {
  size: 'medium',
})

const resolvedSrc = computed(() => {
  if (props.size === 'large') return props.src
  return avatarVariantUrl(props.src)
})

function avatarVariantUrl(src: string): string {
  try {
    const url = new URL(src, window.location.origin)
    const staticMatch = url.pathname.match(/^(\/static\/pic\/(?:[1-8]|default-avatar))\.webp$/)
    if (staticMatch) {
      url.pathname = `${staticMatch[1]}_medium.webp`
      return formatAvatarUrl(src, url)
    }

    const match = url.pathname.match(/^(.*\/)avatar(\.[^/.]+)$/)
    if (!match) return src

    url.pathname = `${match[1]}avatar_medium${match[2]}`
    return formatAvatarUrl(src, url)
  } catch {
    return src
  }
}

function formatAvatarUrl(src: string, url: URL): string {
  if (!src.startsWith('http://') && !src.startsWith('https://')) {
    return `${url.pathname}${url.search}${url.hash}`
  }
  return url.toString()
}
</script>

<template>
  <img :src="resolvedSrc" :alt="alt">
</template>
