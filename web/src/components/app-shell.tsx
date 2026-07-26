import { Link } from '@tanstack/react-router'
import type { ReactNode } from 'react'

import { LaivanWordmark } from './laivan-wordmark'

type AppShellProps = {
  children: ReactNode
}

export function AppShell({ children }: AppShellProps) {
  return (
    <div className="px-safe-left pt-safe-top pr-safe-right pb-safe-bottom bg-background text-foreground min-h-dvh">
      <div className="max-w-content mx-auto w-full px-5 py-5 sm:px-8 sm:py-7 lg:px-10">
        <header className="mb-7 flex min-h-11 items-center sm:mb-10">
          <Link to="/" aria-label="Laivan home" className="focus-ring rounded-control">
            <LaivanWordmark />
          </Link>
        </header>
        <main>{children}</main>
      </div>
    </div>
  )
}
