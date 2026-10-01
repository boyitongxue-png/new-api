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

import { useStatus } from '@/hooks/use-status'
import { useTopNavLinks, type TopNavLink } from '@/hooks/use-top-nav-links'

import { LandingCtaGroup } from './landing-primitives'

const BRAND_NAME = 'API MIX'
const WORDMARK_WORDS = ['API', 'MIX']

interface LandingFooterProps {
  isAuthenticated: boolean
  host: string
}

function FooterNavLink(props: { link: TopNavLink }) {
  if (props.link.external) {
    return (
      <a href={props.link.href} target='_blank' rel='noopener noreferrer'>
        {props.link.title}
      </a>
    )
  }
  if (props.link.requiresAuth) {
    return (
      <Link to='/sign-in' search={{ redirect: props.link.href }}>
        {props.link.title}
      </Link>
    )
  }
  return <Link to={props.link.href}>{props.link.title}</Link>
}

export function LandingFooter(props: LandingFooterProps) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const dynamicLinks = useTopNavLinks()
  const navLinks: TopNavLink[] =
    dynamicLinks.length > 0
      ? dynamicLinks.filter((link) => !link.disabled)
      : [
          { title: t('Home'), href: '/' },
          { title: t('Model Square'), href: '/pricing' },
          { title: t('Rankings'), href: '/rankings' },
          { title: t('Console'), href: '/dashboard' },
        ]
  if (status?.user_agreement_enabled) {
    navLinks.push({ title: t('User Agreement'), href: '/user-agreement' })
  }
  if (status?.privacy_policy_enabled) {
    navLinks.push({ title: t('Privacy Policy'), href: '/privacy-policy' })
  }

  return (
    <footer className='amx-foot' aria-labelledby='amx-foot-title'>
      <div className='amx-wrap amx-foot-cta' data-reveal>
        <h2 className='amx-foot-title' id='amx-foot-title'>
          <span className='amx-ln'>
            <span className='amx-ln-i'>{t('home.landing.footTitleLine1')}</span>
          </span>
          <span className='amx-ln'>
            <span
              className='amx-ln-i'
              style={{ '--i': 1 } as CSSProperties}
            >
              {t('home.landing.footTitleBefore')}
              <em>{t('home.landing.footTitleEmphasis')}</em>
              {t('home.landing.footTitleAfter')}
            </span>
          </span>
        </h2>
        <LandingCtaGroup
          isAuthenticated={props.isAuthenticated}
          className='amx-rv'
        />
      </div>
      <div className='amx-wrap amx-foot-meta'>
        <span className='amx-mono'>
          © {BRAND_NAME}
          {props.host ? ` · ${props.host}` : ''}
        </span>
        <nav
          className='amx-foot-links'
          aria-label={t('home.landing.footNavLabel')}
        >
          {navLinks.map((link) => (
            <FooterNavLink key={link.href} link={link} />
          ))}
        </nav>
        <span className='amx-mono'>{t('home.landing.footTagline')}</span>
      </div>
      <div className='amx-wrap'>
        <div className='amx-wm' aria-hidden='true'>
          {WORDMARK_WORDS.map((word, wordIndex) => (
            <span key={word} className='amx-wm-word'>
              {wordIndex > 0 ? <span className='amx-wm-sp' /> : null}
              {word.split('').map((letter, letterIndex) => (
                <span key={`${word}-${letterIndex}`} className='amx-wm-l'>
                  <span>{letter}</span>
                </span>
              ))}
            </span>
          ))}
        </div>
      </div>
    </footer>
  )
}
