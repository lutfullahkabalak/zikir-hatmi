const rawApiBase = (import.meta.env.VITE_API_BASE ?? '').toString().trim()

const apiBase = rawApiBase.replace(/\/+$/, '')

export const API_BASE = apiBase

export const apiUrl = (path: string): string => {
  const normalized = path.startsWith('/') ? path : `/${path}`
  return apiBase ? `${apiBase}${normalized}` : normalized
}

export const wsUrl = (path: string): string => {
  const normalized = path.startsWith('/') ? path : `/${path}`

  if (apiBase) {
    const wsBase = apiBase.replace(/^http(s?):\/\//i, (_match: string, s: string) => `ws${s}://`)
    return `${wsBase}${normalized}`
  }

  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  return `${protocol}://${window.location.host}${normalized}`
}
