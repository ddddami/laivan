import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider, createRouter } from '@tanstack/react-router'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

import bricolageFontUrl from '@fontsource-variable/bricolage-grotesque/files/bricolage-grotesque-latin-wght-normal.woff2?url'
import geistFontUrl from '@fontsource-variable/geist/files/geist-latin-wght-normal.woff2?url'
import geistMonoFontUrl from '@fontsource-variable/geist-mono/files/geist-mono-latin-wght-normal.woff2?url'

import './styles.css'
import { createAppQueryClient } from './query-client'
import { routeTree } from './routeTree.gen'

const queryClient = createAppQueryClient()

for (const href of [bricolageFontUrl, geistFontUrl, geistMonoFontUrl]) {
  const preload = document.createElement('link')
  preload.rel = 'preload'
  preload.as = 'font'
  preload.type = 'font/woff2'
  preload.crossOrigin = 'anonymous'
  preload.href = href
  document.head.append(preload)
}

const router = createRouter({
  routeTree,
  context: { queryClient },
  scrollRestoration: true,
  defaultPreload: 'intent',
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

const root = document.getElementById('root')

if (!root) {
  throw new Error('Root element was not found')
}

createRoot(root).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
)
