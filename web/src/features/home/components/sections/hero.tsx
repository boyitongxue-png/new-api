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
import { ArrowRight, BookOpen, Check } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { useStatus } from '@/hooks/use-status'

import { HeroTerminalDemo } from '../hero-terminal-demo'

interface HeroProps {
  className?: string
  isAuthenticated?: boolean
}

export function Hero(props: HeroProps) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const docsUrl =
    (status?.docs_link as string | undefined) || 'https://docs.newapi.pro'

  const renderDocsButton = () => {
    const isExternal = docsUrl.startsWith('http')
    if (isExternal) {
      return (
        <Button
          variant='outline'
          className='group inline-flex h-11 items-center gap-2 rounded-lg border-[#344158] bg-white/[0.03] px-5 text-sm font-semibold text-[#24334d] hover:border-[#5d7196] hover:bg-white/[0.07] dark:border-[#344158] dark:bg-white/[0.03] dark:text-[#dce6f7]'
          render={
            <a href={docsUrl} target='_blank' rel='noopener noreferrer' />
          }
        >
          <BookOpen className='size-4 text-[#52627e] transition-colors duration-200 group-hover:text-[#24334d] dark:text-[#91a4c1] dark:group-hover:text-[#dce6f7]' />
          <span>{t('Read documentation')}</span>
        </Button>
      )
    }
    return (
      <Button
        variant='outline'
        className='group inline-flex h-11 items-center gap-2 rounded-lg border-[#344158] bg-white/[0.03] px-5 text-sm font-semibold text-[#24334d] hover:border-[#5d7196] hover:bg-white/[0.07] dark:border-[#344158] dark:bg-white/[0.03] dark:text-[#dce6f7]'
        render={<Link to={docsUrl} />}
      >
        <BookOpen className='size-4 text-[#52627e] transition-colors duration-200 group-hover:text-[#24334d] dark:text-[#91a4c1] dark:group-hover:text-[#dce6f7]' />
        <span>{t('Read documentation')}</span>
      </Button>
    )
  }

  return (
    <section className='relative z-10 overflow-hidden bg-[#edf4fc] px-6 pt-24 pb-16 text-[#121b2c] md:pb-20 lg:pt-24 lg:pb-16 dark:bg-[#090d16] dark:text-[#f5f7fa]'>
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 -z-10 opacity-70 dark:opacity-100'
        style={{
          backgroundImage:
            'linear-gradient(rgba(91,108,140,.09) 1px, transparent 1px), linear-gradient(90deg, rgba(91,108,140,.09) 1px, transparent 1px)',
          backgroundSize: '72px 72px',
        }}
      />

      <div className='relative mx-auto grid w-full max-w-7xl grid-cols-1 items-center gap-12 lg:grid-cols-[0.9fr_1.1fr] lg:gap-16'>
        <div className='flex max-w-[620px] flex-col items-start text-left'>
          <div
            className='landing-animate-fade-up mb-6 inline-flex items-center gap-2 rounded-lg border border-[#2f80ed]/40 bg-[#2f80ed]/10 px-3 py-1.5 text-xs font-semibold text-[#2563b8] opacity-0 dark:text-[#67a7ff]'
            style={{ animationDelay: '0ms' }}
          >
            <span className='size-1.5 rounded-full bg-[#44d7b6]' />
            <span>{t('AI Gateway for AI')}</span>
          </div>

          <h1
            className='landing-animate-fade-up max-w-[11ch] text-[clamp(2.7rem,6.2vw,5.5rem)] leading-[0.98] font-semibold tracking-[0] text-balance opacity-0'
            style={{ animationDelay: '60ms' }}
          >
            {t('One Gateway. All Intelligence.')}
          </h1>
          <p
            className='landing-animate-fade-up mt-7 max-w-[560px] text-base leading-8 text-[#52627e] opacity-0 sm:text-lg dark:text-[#8d9bb3]'
            style={{ animationDelay: '140ms' }}
          >
            {t(
              'A unified API, smart routing, and real-time observability let your team build with global AI models through one reliable interface.'
            )}
          </p>

          <div
            className='landing-animate-fade-up mt-8 flex flex-wrap items-center gap-3 opacity-0'
            style={{ animationDelay: '220ms' }}
          >
            {props.isAuthenticated ? (
              <>
                <Button
                  className='group h-11 rounded-lg bg-[#2f80ed] px-5 text-sm font-semibold text-white hover:bg-[#4392ff] active:translate-y-px'
                  render={<Link to='/dashboard' />}
                >
                  {t('Go to Dashboard')}
                  <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
                </Button>
                {renderDocsButton()}
              </>
            ) : (
              <>
                <Button
                  className='group h-11 rounded-lg bg-[#2f80ed] px-5 text-sm font-semibold text-white hover:bg-[#4392ff] active:translate-y-px'
                  render={<Link to='/sign-up' />}
                >
                  {t('Build with API MIX')}
                  <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
                </Button>
                {renderDocsButton()}
              </>
            )}
          </div>

          <div
            className='landing-animate-fade-up mt-9 flex flex-wrap gap-x-6 gap-y-2 text-xs text-[#62718b] opacity-0 dark:text-[#71809a]'
            style={{ animationDelay: '300ms' }}
          >
            {[
              'OpenAI-compatible API',
              'Usage-based billing',
              'Measured availability',
            ].map((item) => (
              <span key={item} className='flex items-center gap-2'>
                <Check className='size-3.5 text-[#1aa884] dark:text-[#44d7b6]' />
                {t(item)}
              </span>
            ))}
          </div>
        </div>

        <div
          className='landing-animate-fade-up flex w-full justify-center opacity-0'
          style={{ animationDelay: '180ms' }}
        >
          <HeroTerminalDemo className='w-full lg:mt-0' />
        </div>
      </div>
    </section>
  )
}
