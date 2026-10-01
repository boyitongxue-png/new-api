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
import type { CSSProperties, ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

const MARQUEE_SET_COUNT = 4

type Capability = {
  title: string
  description: string
  icon: ReactNode
}

function delayStyle(index: number): CSSProperties {
  return { '--i': index } as CSSProperties
}

function CapabilityIcon(props: { index: number; children: ReactNode }) {
  return (
    <svg
      className='amx-cap-icon'
      style={delayStyle(props.index)}
      viewBox='0 0 24 24'
      fill='none'
      stroke='currentColor'
      strokeWidth='1.5'
      strokeLinecap='round'
      strokeLinejoin='round'
      aria-hidden='true'
    >
      {props.children}
    </svg>
  )
}

export function LandingBand() {
  const { t } = useTranslation()
  const marqueeItems = [
    { text: 'Chat Completions', outline: false },
    { text: 'Responses', outline: true },
    { text: 'Messages', outline: false },
    { text: 'generateContent', outline: true },
    { text: t('home.landing.marquee.compatible'), outline: false },
    { text: 'Claude', outline: true },
    { text: 'Gemini', outline: false },
    { text: 'One Key', outline: true },
  ]
  const capabilities: Capability[] = [
    {
      title: t('home.landing.cap.gateway.title'),
      description: t('home.landing.cap.gateway.description'),
      icon: (
        <CapabilityIcon index={0}>
          <rect x='3.5' y='3.5' width='17' height='17' rx='2' pathLength={1} />
          <path d='M9 9h6v6H9z' pathLength={1} />
        </CapabilityIcon>
      ),
    },
    {
      title: t('home.landing.cap.routing.title'),
      description: t('home.landing.cap.routing.description'),
      icon: (
        <CapabilityIcon index={1}>
          <circle cx='6' cy='5.5' r='2.5' pathLength={1} />
          <circle cx='18' cy='18.5' r='2.5' pathLength={1} />
          <path
            d='M8.5 5.5H14a3.75 3.75 0 0 1 0 7.5h-4a3 3 0 0 0 0 5.5h5.5'
            pathLength={1}
          />
        </CapabilityIcon>
      ),
    },
    {
      title: t('home.landing.cap.observability.title'),
      description: t('home.landing.cap.observability.description'),
      icon: (
        <CapabilityIcon index={2}>
          <path d='M3 20.5h18' pathLength={1} />
          <path d='M6.5 17v-5' pathLength={1} />
          <path d='M11.5 17V5.5' pathLength={1} />
          <path d='M16.5 17V9' pathLength={1} />
        </CapabilityIcon>
      ),
    },
    {
      title: t('home.landing.cap.access.title'),
      description: t('home.landing.cap.access.description'),
      icon: (
        <CapabilityIcon index={3}>
          <path
            d='M12 3l8 3v6c0 4.5-3.4 8-8 9-4.6-1-8-4.5-8-9V6z'
            pathLength={1}
          />
          <path d='M9 12l2 2 4-4' pathLength={1} />
        </CapabilityIcon>
      ),
    },
  ]

  return (
    <section className='amx-band' aria-labelledby='amx-band-title'>
      <div className='amx-marquee' aria-hidden='true'>
        <div className='amx-mq-track'>
          {Array.from({ length: MARQUEE_SET_COUNT }, (_, setIndex) => (
            <div key={setIndex} className='amx-mq-set'>
              {marqueeItems.map((item) => (
                <span
                  key={item.text}
                  className={cn('amx-mq-item', item.outline && 'is-outline')}
                >
                  {item.text}
                </span>
              ))}
            </div>
          ))}
        </div>
      </div>
      <div className='amx-wrap amx-band-inner'>
        <div className='amx-band-head' data-reveal>
          <h2 className='amx-band-title' id='amx-band-title'>
            <span className='amx-ln'>
              <span className='amx-ln-i'>
                {t('home.landing.bandTitleLine1')}
              </span>
            </span>
            <span className='amx-ln'>
              <span className='amx-ln-i' style={delayStyle(1)}>
                {t('home.landing.bandTitleLine2')}
              </span>
            </span>
          </h2>
          <p className='amx-mono amx-rv' style={delayStyle(3)}>
            {t('home.landing.bandMeta')} / 04
          </p>
        </div>
        <ul className='amx-caps' data-reveal>
          {capabilities.map((capability, index) => (
            <li
              key={capability.title}
              className='amx-cap amx-rv'
              style={delayStyle(index)}
            >
              <div className='amx-cap-top'>
                {capability.icon}
                <span className='amx-mono'>
                  {String(index + 1).padStart(2, '0')}
                </span>
              </div>
              <h3>{capability.title}</h3>
              <p>{capability.description}</p>
            </li>
          ))}
        </ul>
      </div>
    </section>
  )
}
