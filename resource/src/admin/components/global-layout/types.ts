import type { Component } from 'vue'

export interface LayoutHeaderProps {
  title: string
  description: string
}

export interface TwoColAsideNavItem {
  title: string
  url: string
  icon?: Component
}
