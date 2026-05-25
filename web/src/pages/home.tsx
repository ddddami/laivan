export function Home() {
  return (
    <main className="px-safe-left pt-safe-top pr-safe-right pb-safe-bottom min-h-dvh bg-[#f7f4eb] text-[#10231d]">
      <div className="mx-auto flex min-h-dvh w-full max-w-7xl flex-col px-5 py-5 sm:px-8 lg:px-10 lg:py-8">
        <nav className="flex items-center justify-between gap-6" aria-label="Main navigation">
          <a className="text-lg font-black tracking-tight" href="/">
            Laivan
          </a>
          <a
            className="rounded-full bg-[#10231d] px-4 py-2 text-sm font-semibold whitespace-nowrap text-[#f7f4eb] shadow-sm transition hover:bg-[#1b3a31]"
            href="https://wa.me/"
          >
            Join waitlist
          </a>
        </nav>

        <section className="grid flex-1 items-center gap-10 py-12 sm:py-16 lg:grid-cols-[minmax(0,1fr)_minmax(360px,0.78fr)] lg:gap-16 lg:py-12 xl:gap-24">
          <div className="max-w-3xl">
            <p className="mb-4 w-fit rounded-full border border-[#d8c7a3] bg-white/70 px-3 py-1 text-sm font-semibold text-[#476256]">
              FUTA accommodation, less chaos
            </p>
            <h1 className="text-5xl leading-[0.95] font-black tracking-[-0.06em] text-balance sm:text-6xl lg:text-7xl xl:text-8xl">
              Find rooms faster without trusting random posts blindly.
            </h1>
            <p className="mt-6 max-w-2xl text-lg leading-8 text-[#476256] sm:text-xl">
              Laivan is building a mobile-first way for students to discover accommodation, compare
              agent signals, and move into WhatsApp with better context.
            </p>
            <div className="mt-8 grid gap-3 sm:flex">
              <a
                className="rounded-2xl bg-[#e78b3a] px-5 py-4 text-center text-base font-black text-[#10231d] shadow-sm transition hover:bg-[#f09a4d]"
                href="https://wa.me/"
              >
                Start with FUTA
              </a>
              <a
                className="rounded-2xl border border-[#d8c7a3] bg-white/70 px-5 py-4 text-center text-base font-black text-[#10231d] transition hover:bg-white"
                href="mailto:hello@laivan.local"
              >
                Talk to us
              </a>
            </div>
          </div>

          <aside className="mx-auto w-full max-w-md rounded-[2rem] border border-[#d8c7a3] bg-white p-3 shadow-[0_24px_80px_rgba(16,35,29,0.12)] sm:p-4 lg:mx-0 lg:max-w-none lg:justify-self-end">
            <div className="rounded-[1.5rem] bg-[#10231d] p-4 text-[#f7f4eb] sm:p-5">
              <div className="flex items-center justify-between gap-4 text-sm text-[#bed1c6]">
                <span>South Gate</span>
                <span>3 agents active</span>
              </div>
              <div className="mt-24 rounded-3xl bg-[#f7f4eb] p-4 text-[#10231d] sm:mt-36 lg:mt-44">
                <p className="text-xs font-bold tracking-[0.18em] text-[#476256] uppercase">
                  Agent offer
                </p>
                <h2 className="mt-2 text-2xl font-black tracking-tight">Self-con near FUTA</h2>
                <dl className="mt-4 grid grid-cols-3 gap-3 text-sm">
                  <div>
                    <dt className="text-[#476256]">Trust</dt>
                    <dd className="font-black">Rising</dd>
                  </div>
                  <div>
                    <dt className="text-[#476256]">Media</dt>
                    <dd className="font-black">12</dd>
                  </div>
                  <div>
                    <dt className="text-[#476256]">Reply</dt>
                    <dd className="font-black">Fast</dd>
                  </div>
                </dl>
              </div>
            </div>
          </aside>
        </section>
      </div>
    </main>
  )
}
