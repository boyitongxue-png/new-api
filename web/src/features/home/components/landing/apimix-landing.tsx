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
import { useRef } from 'react'

import { cn } from '@/lib/utils'

import { useLandingIntroHold, useLandingMotion } from '../../hooks'
import { LandingBand } from './landing-band'
import { LandingEndpoints } from './landing-endpoints'
import { LandingFooter } from './landing-footer'
import { LandingHero } from './landing-hero'

interface ApiMixLandingProps {
  isAuthenticated: boolean
}

/**
 * Built-in API MIX landing page (black × acid lime), rendered when no custom
 * home page content is configured. Styles live in `src/styles/apimix.css`
 * under the `.amx-home` scope; motion is wired by `useLandingMotion`.
 */
export function ApiMixLanding(props: ApiMixLandingProps) {
  const rootRef = useRef<HTMLDivElement>(null)
  const holding = useLandingIntroHold()
  useLandingMotion(rootRef)
  const host = typeof window === 'undefined' ? '' : window.location.host

  return (
    <div ref={rootRef} className={cn('amx-home', holding && 'is-holding')}>
      <main>
        <LandingHero isAuthenticated={props.isAuthenticated} host={host} />
        <LandingEndpoints />
        <LandingBand />
      </main>
      <LandingFooter isAuthenticated={props.isAuthenticated} host={host} />
    </div>
  )
}
