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
import type { CSSProperties } from 'react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

const ZERO_WIDTH_SPACE = '​'

type Endpoint = {
  path: string
  label: string
  live?: boolean
}

function revealDelay(index: number): CSSProperties {
  return { '--i': index } as CSSProperties
}

export function LandingEndpoints() {
  const { t } = useTranslation()
  const endpoints: Endpoint[] = [
    { path: '/v1/chat/completions', label: t('home.landing.ep.chat') },
    {
      path: '/v1/responses',
      label: t('home.landing.ep.responses'),
      live: true,
    },
    { path: '/v1/messages', label: t('home.landing.ep.messages') },
    {
      // Zero-width space lets the long Gemini path wrap after the slash.
      path: `/v1beta/models/${ZERO_WIDTH_SPACE}*:generateContent`,
      label: t('home.landing.ep.gemini'),
    },
  ]

  return (
    <section className='amx-endpoints' aria-labelledby='amx-ep-title'>
      <div className='amx-wrap'>
        <div className='amx-sec-head' data-reveal>
          <h2 className='amx-h2' id='amx-ep-title'>
            <span className='amx-ln'>
              <span className='amx-ln-i'>
                {t('home.landing.endpointsTitle')}
              </span>
            </span>
          </h2>
          <p className='amx-mono amx-sec-meta amx-rv' style={revealDelay(2)}>
            <span>{t('home.landing.endpointsMeta')}</span>
            <b>/ 04</b>
          </p>
        </div>
        <ol className='amx-ep-list' data-reveal>
          {endpoints.map((endpoint, index) => {
            const readablePath = endpoint.path.replace(ZERO_WIDTH_SPACE, '')
            const ariaLabel = endpoint.live
              ? t('home.landing.ep.ariaLive', {
                  path: readablePath,
                  label: endpoint.label,
                })
              : t('home.landing.ep.aria', {
                  path: readablePath,
                  label: endpoint.label,
                })
            return (
              <li
                key={endpoint.path}
                className='amx-rv'
                style={revealDelay(index + 1)}
              >
                <Link
                  to='/pricing'
                  className={cn('amx-ep', endpoint.live && 'is-active')}
                  aria-label={ariaLabel}
                >
                  <span className='amx-ep-idx' aria-hidden='true'>
                    {String(index + 1).padStart(2, '0')}
                  </span>
                  <span
                    className='amx-ep-path'
                    data-scramble={endpoint.path}
                    aria-hidden='true'
                  >
                    {endpoint.path}
                  </span>
                  {endpoint.live ? (
                    <span className='amx-ep-live' aria-hidden='true'>
                      <i className='amx-dot' />
                      <span>
                        200<span className='amx-ok-t'> OK</span>
                      </span>
                    </span>
                  ) : (
                    <span aria-hidden='true' />
                  )}
                  <span className='amx-ep-label' aria-hidden='true'>
                    {endpoint.label}
                  </span>
                  <svg
                    className='amx-ep-go'
                    viewBox='0 0 24 24'
                    fill='none'
                    stroke='currentColor'
                    strokeWidth='1.5'
                    strokeLinecap='round'
                    strokeLinejoin='round'
                    aria-hidden='true'
                  >
                    <path d='M3 12h17M14 6l6 6-6 6' />
                  </svg>
                </Link>
              </li>
            )
          })}
        </ol>
      </div>
    </section>
  )
}
