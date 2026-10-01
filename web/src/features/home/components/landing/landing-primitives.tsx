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

import { cn } from '@/lib/utils'

function ArrowPath() {
  return (
    <svg
      viewBox='0 0 16 16'
      fill='none'
      stroke='currentColor'
      strokeWidth='1.6'
      strokeLinecap='round'
      strokeLinejoin='round'
    >
      <path d='M2.5 8h11M9 3.5 13.5 8 9 12.5' />
    </svg>
  )
}

/** Two stacked arrows that swap on hover (see `.amx-arr` in apimix.css). */
export function LandingArrow() {
  return (
    <span className='amx-arr' aria-hidden='true'>
      <ArrowPath />
      <ArrowPath />
    </span>
  )
}

export function LandingCheckIcon() {
  return (
    <svg
      viewBox='0 0 12 12'
      fill='none'
      stroke='currentColor'
      strokeWidth='1.6'
      strokeLinecap='round'
      strokeLinejoin='round'
      aria-hidden='true'
    >
      <path d='M2 6.3 4.7 9 10 3' />
    </svg>
  )
}

interface LandingCtaGroupProps {
  isAuthenticated: boolean
  className?: string
}

/**
 * Primary + secondary landing CTAs. Mirrors the auth-aware routing of the
 * previous hero: signed-in visitors go straight to the console, everyone else
 * is sent to sign in first.
 */
export function LandingCtaGroup(props: LandingCtaGroupProps) {
  const { t } = useTranslation()
  return (
    <div className={cn('amx-cta-group', props.className)}>
      <Link
        to={props.isAuthenticated ? '/dashboard' : '/sign-in'}
        className='amx-btn amx-btn-primary amx-magnetic'
      >
        <span className='amx-btn-label'>
          {t('home.landing.ctaConsole')}
          <LandingArrow />
        </span>
      </Link>
      <Link to='/pricing' className='amx-btn amx-btn-ghost amx-magnetic'>
        <span className='amx-btn-label'>{t('home.landing.ctaPricing')}</span>
      </Link>
    </div>
  )
}
