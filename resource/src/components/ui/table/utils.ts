import type { Ref } from "vue"

type Updater<T> = T | ((old: T) => T)

function isFunction<T>(value: Updater<T>): value is (old: T) => T {
  return typeof value === "function"
}

export function valueUpdater<T>(updaterOrValue: Updater<T>, ref: Ref<T>) {
  ref.value = isFunction(updaterOrValue)
    ? updaterOrValue(ref.value)
    : updaterOrValue
}
