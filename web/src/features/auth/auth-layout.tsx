/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'
import { DEFAULT_LOGO } from '@/lib/constants'

type AuthLayoutProps = {
  children: React.ReactNode
}

function AuthBrandLink() {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()
  const hasCustomLogo = Boolean(logo) && logo !== DEFAULT_LOGO

  // Admin-configured logos win; the shipped default becomes the lime mark.
  let brandMark: React.ReactNode
  if (loading) {
    brandMark = <Skeleton className='size-3 rounded-none' />
  } else if (hasCustomLogo) {
    brandMark = (
      <img
        src={logo}
        alt={t('Logo')}
        className='size-7 shrink-0 rounded-md object-contain'
      />
    )
  } else {
    brandMark = (
      <span
        aria-hidden='true'
        className='block size-[11px] shrink-0 bg-[var(--brand-lime,var(--primary))] transition-transform duration-500 group-hover:rotate-90 motion-reduce:transition-none'
      />
    )
  }

  return (
    <Link
      to='/'
      className='group focus-visible:ring-ring/60 inline-flex min-w-0 items-center gap-2.5 rounded-md transition-opacity outline-none hover:opacity-90 focus-visible:ring-2'
    >
      {brandMark}
      {loading ? (
        <Skeleton className='h-5 w-24' />
      ) : (
        <span className='truncate text-lg font-bold tracking-[-0.02em]'>
          {systemName}
        </span>
      )}
    </Link>
  )
}

/** Brand side of the split auth layout (below the form on phones). */
function AuthBrandPanel() {
  const { t } = useTranslation()
  const host = typeof window === 'undefined' ? '' : window.location.host
  const checks = [
    t('home.landing.check.compatible'),
    t('home.landing.check.billing'),
    t('home.landing.check.stable'),
  ]

  return (
    <section
      aria-labelledby='auth-brand-title'
      className='border-border relative isolate flex flex-col justify-between gap-10 overflow-hidden border-t px-4 py-10 sm:px-8 lg:min-h-svh lg:border-t-0 lg:border-r lg:px-12 lg:py-10'
    >
      <span className='amx-auth-mix' aria-hidden='true'>
        MIX
      </span>
      <div className='hidden lg:block'>
        <AuthBrandLink />
      </div>
      <div className='max-w-xl'>
        <p className='amx-auth-eyebrow text-brand'>
          {t('home.landing.eyebrow')}
        </p>
        <h2 id='auth-brand-title' className='amx-auth-title'>
          <span className='block'>{t('home.landing.titleLine1')}</span>
          <span className='block'>
            {t('home.landing.titleLine2Before')}
            <span className='amx-auth-hl'>
              {t('home.landing.titleHighlight')}
            </span>
            {t('home.landing.titleLine2After')}
          </span>
        </h2>
        <p className='font-apimix-cn text-muted-foreground mt-6 max-w-[26em] text-[15px] leading-relaxed sm:text-base'>
          {t('Sign in to manage API keys, track usage and top up online.')}
        </p>
        <ul
          className='font-apimix-cn mt-6 flex flex-wrap gap-x-6 gap-y-2 text-sm sm:text-[15px]'
          aria-label={t('home.landing.checksLabel')}
        >
          {checks.map((label) => (
            <li key={label} className='inline-flex items-center gap-2'>
              <svg
                viewBox='0 0 12 12'
                fill='none'
                stroke='currentColor'
                strokeWidth='1.6'
                strokeLinecap='round'
                strokeLinejoin='round'
                aria-hidden='true'
                className='text-brand size-3 shrink-0'
              >
                <path d='M2 6.3 4.7 9 10 3' />
              </svg>
              {label}
            </li>
          ))}
        </ul>
      </div>
      <p className='font-apimix-mono text-muted-foreground text-xs tracking-[0.08em]'>
        © API MIX{host ? ` · ${host}` : ''}
      </p>
    </section>
  )
}

/**
 * Split auth layout: brand panel on the left and the (unchanged) auth form
 * on the right on desktop; on phones the form comes first and the brand
 * panel follows below it.
 */
export function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className='bg-background relative grid min-h-svh max-w-none lg:grid-cols-2'>
      <div className='order-1 flex min-h-svh flex-col lg:order-2'>
        <div className='px-4 pt-5 sm:px-8 sm:pt-8 lg:hidden'>
          <AuthBrandLink />
        </div>
        <div className='flex flex-1 items-center justify-center px-4 py-8 sm:px-8'>
          <div className='bg-card text-card-foreground border-border flex w-full max-w-[440px] flex-col justify-center space-y-2 rounded-2xl border p-5 shadow-sm sm:p-8'>
            {children}
          </div>
        </div>
      </div>
      <div className='order-2 lg:order-1'>
        <AuthBrandPanel />
      </div>
    </div>
  )
}
