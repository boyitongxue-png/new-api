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
import { useTranslation } from 'react-i18next'

import { LandingCheckIcon, LandingCtaGroup } from './landing-primitives'

interface LandingHeroProps {
  isAuthenticated: boolean
  host: string
}

function HeroMixWordmark() {
  return (
    <div className='amx-hero-mix' aria-hidden='true'>
      <svg viewBox='0 0 1040 430' focusable='false'>
        <defs>
          <linearGradient id='amx-sheen-grad' x1='0' y1='0' x2='1' y2='0'>
            <stop offset='0' stopColor='#fff' stopOpacity='0' />
            <stop offset='.5' stopColor='#fff' stopOpacity='.9' />
            <stop offset='1' stopColor='#fff' stopOpacity='0' />
          </linearGradient>
          <mask
            id='amx-sheen-mask'
            maskUnits='userSpaceOnUse'
            x='-200'
            y='-100'
            width='1500'
            height='700'
          >
            <rect
              className='amx-sheen-rect'
              x='0'
              y='-100'
              width='380'
              height='700'
              fill='url(#amx-sheen-grad)'
            />
          </mask>
        </defs>
        <text
          className='amx-mix-text amx-mix-base'
          x='1030'
          y='414'
          textAnchor='end'
        >
          MIX
        </text>
        <text
          className='amx-mix-text amx-mix-sheen'
          x='1030'
          y='414'
          textAnchor='end'
          mask='url(#amx-sheen-mask)'
        >
          MIX
        </text>
      </svg>
    </div>
  )
}

export function LandingHero(props: LandingHeroProps) {
  const { t } = useTranslation()
  const highlight = t('home.landing.titleHighlight')
  const checks = [
    t('home.landing.check.compatible'),
    t('home.landing.check.billing'),
    t('home.landing.check.stable'),
  ]

  return (
    <section className='amx-hero' aria-labelledby='amx-hero-title'>
      <HeroMixWordmark />
      <div className='amx-hero-glow' aria-hidden='true' />

      <div className='amx-wrap amx-hero-inner'>
        <p className='amx-eyebrow'>{t('home.landing.eyebrow')}</p>
        <h1 className='amx-hero-title' id='amx-hero-title'>
          <span className='amx-ln'>
            <span className='amx-ln-i'>{t('home.landing.titleLine1')}</span>
          </span>
          <span className='amx-ln'>
            <span className='amx-ln-i'>
              {t('home.landing.titleLine2Before')}
              <span className='amx-hl'>
                {highlight}
                <span className='amx-hl-fill' aria-hidden='true'>
                  <span>{highlight}</span>
                </span>
              </span>
              {t('home.landing.titleLine2After')}
            </span>
          </span>
        </h1>
        <div className='amx-hero-row'>
          <p className='amx-lede'>{t('home.landing.lede')}</p>
          <LandingCtaGroup isAuthenticated={props.isAuthenticated} />
        </div>
      </div>

      <div className='amx-wrap amx-hero-foot'>
        <ul className='amx-checks' aria-label={t('home.landing.checksLabel')}>
          {checks.map((label) => (
            <li key={label}>
              <LandingCheckIcon />
              {label}
            </li>
          ))}
        </ul>
        {props.host ? <span className='amx-mono'>{props.host}</span> : null}
      </div>
    </section>
  )
}
