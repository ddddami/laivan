import { Cancel01Icon } from '@hugeicons/core-free-icons'
import { Dialog } from 'radix-ui'
import type { ReactNode } from 'react'

import { IconButton } from './icon-button'

type SheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  children: ReactNode
}

export function Sheet({ open, onOpenChange, title, description, children }: SheetProps) {
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="duration-standard fixed inset-0 z-40 bg-black/25 backdrop-blur-sm transition-opacity data-[state=closed]:opacity-0 data-[state=open]:opacity-100" />
        <Dialog.Content className="bg-background rounded-t-sheet duration-standard fixed inset-x-0 bottom-0 z-50 flex max-h-[92dvh] flex-col shadow-[0_-16px_50px_rgba(15,15,15,0.14)] transition-transform focus:outline-none data-[state=closed]:translate-y-full data-[state=open]:translate-y-0 sm:inset-y-0 sm:right-0 sm:left-auto sm:max-h-none sm:w-full sm:max-w-md sm:rounded-none sm:data-[state=closed]:translate-x-full sm:data-[state=closed]:translate-y-0 sm:data-[state=open]:translate-x-0">
          <div className="border-border flex shrink-0 items-center justify-between border-b px-5 py-3">
            <div>
              <Dialog.Title className="font-display text-foreground text-base font-bold tracking-[-0.02em]">
                {title}
              </Dialog.Title>
              <Dialog.Description className="font-body text-faint mt-0.5 text-xs">
                {description}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <IconButton icon={Cancel01Icon} label={`Close ${title.toLowerCase()}`} />
            </Dialog.Close>
          </div>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
