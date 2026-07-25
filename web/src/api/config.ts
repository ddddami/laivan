export type PublicConfig = {
  apiBaseUrl: string
}

export function createPublicConfig(value: string | undefined): PublicConfig {
  const apiBaseUrl = value?.trim()
  if (!apiBaseUrl) {
    return { apiBaseUrl: '' }
  }

  let url: URL
  try {
    url = new URL(apiBaseUrl)
  } catch {
    throw new Error('VITE_API_BASE_URL must be an absolute URL')
  }
  if (url.protocol !== 'http:' && url.protocol !== 'https:') {
    throw new Error('VITE_API_BASE_URL must use http or https')
  }
  if (url.search || url.hash) {
    throw new Error('VITE_API_BASE_URL must not include a query string or fragment')
  }

  return { apiBaseUrl: apiBaseUrl.replace(/\/+$/, '') }
}

export const publicConfig = createPublicConfig(import.meta.env.VITE_API_BASE_URL)
