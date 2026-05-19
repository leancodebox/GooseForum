export const motionDurations = {
  instant: 0.12,
  fast: 0.16,
  standard: 0.22,
  comfortable: 0.28,
} as const

export const motionEase = {
  standard: [0.22, 1, 0.36, 1],
  emphasized: [0.16, 1, 0.3, 1],
} as const

export const motionTransitions = {
  fast: {
    duration: motionDurations.fast,
    ease: motionEase.standard,
  },
  standard: {
    duration: motionDurations.standard,
    ease: motionEase.standard,
  },
  comfortable: {
    duration: motionDurations.comfortable,
    ease: motionEase.emphasized,
  },
} as const

export const overlayMotion = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0 },
} as const

export const mobileDrawerMotion = {
  initial: { x: '-100%', opacity: 0.96 },
  animate: { x: 0, opacity: 1 },
  exit: { x: '-100%', opacity: 0.96 },
} as const
