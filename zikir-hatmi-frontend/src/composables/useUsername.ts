import { ref } from 'vue'

const STORAGE_KEY = 'zikir-username'

const readStored = () => {
  if (typeof window === 'undefined') return ''
  return localStorage.getItem(STORAGE_KEY) || ''
}

const username = ref<string>(readStored())

export const useUsername = () => {
  const setUsername = (next: string) => {
    const value = next.trim()
    username.value = value

    if (typeof window === 'undefined') return

    if (value) {
      localStorage.setItem(STORAGE_KEY, value)
    } else {
      localStorage.removeItem(STORAGE_KEY)
    }
  }

  return {
    username,
    setUsername,
  }
}
