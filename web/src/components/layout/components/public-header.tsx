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
import { Link, useNavigate, useRouterState } from '@tanstack/react-router'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Dialog } from '@/components/dialog'
import { LanguageSwitcher } from '@/components/language-switcher'
import { NotificationPopover } from '@/components/notification-popover'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ThemeSwitch } from '@/components/theme-switch'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useNotifications } from '@/hooks/use-notifications'
import { useSystemConfig } from '@/hooks/use-system-config'
import { useTopNavLinks } from '@/hooks/use-top-nav-links'
import { DEFAULT_LOGO } from '@/lib/constants'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

import { defaultTopNavLinks } from '../config/top-nav.config'
import type { TopNavLink } from '../types'
import { HeaderLogo } from './header-logo'

const AUTH_PROMPT_SECONDS = 5

type AuthPromptTarget = {
  title: string
  href: string
}

type PillIndicator = {
  left: number
  width: number
}

/** A nav link is current on its own path and on nested paths below it. */
function isNavLinkActive(pathname: string, href: string): boolean {
  if (pathname === href) return true
  return href !== '/' && pathname.startsWith(`${href}/`)
}

/**
 * Brand mark next to the site name: an admin-provided logo wins; the shipped
 * default logo is replaced by the API MIX lime square.
 */
function PublicHeaderBrandMark(props: {
  customLogo?: React.ReactNode
  systemLogo: string
  logoLoaded: boolean
}) {
  if (props.customLogo) {
    return (
      <span className='flex size-7 shrink-0 items-center justify-center'>
        {props.customLogo}
      </span>
    )
  }
  if (props.systemLogo && props.systemLogo !== DEFAULT_LOGO) {
    return (
      <HeaderLogo
        src={props.systemLogo}
        loading={false}
        logoLoaded={props.logoLoaded}
        className='size-7 shrink-0 rounded-md object-contain'
      />
    )
  }
  return (
    <span
      aria-hidden='true'
      className='block size-[11px] shrink-0 bg-[var(--brand-lime,var(--primary))] transition-transform duration-500 ease-[cubic-bezier(0.34,1.56,0.64,1)] group-hover:scale-110 group-hover:rotate-90 motion-reduce:transition-none'
    />
  )
}

export interface PublicHeaderProps {
  navLinks?: TopNavLink[]
  mobileLinks?: TopNavLink[]
  navContent?: React.ReactNode
  showThemeSwitch?: boolean
  showLanguageSwitcher?: boolean
  logo?: React.ReactNode
  siteName?: string
  homeUrl?: string
  leftContent?: React.ReactNode
  rightContent?: React.ReactNode
  showNavigation?: boolean
  showAuthButtons?: boolean
  showNotifications?: boolean
  className?: string
}

export function PublicHeader(props: PublicHeaderProps) {
  const {
    navLinks = defaultTopNavLinks,
    showThemeSwitch = true,
    showLanguageSwitcher = true,
    logo: customLogo,
    siteName: customSiteName,
    homeUrl = '/',
    showAuthButtons = true,
    showNotifications = true,
  } = props

  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const [scrolled, setScrolled] = useState(false)
  const pillRef = useRef<HTMLElement>(null)
  const [pillIndicator, setPillIndicator] = useState<PillIndicator | null>(
    null
  )
  const [mobileOpen, setMobileOpen] = useState(false)
  const [authPromptTarget, setAuthPromptTarget] =
    useState<AuthPromptTarget | null>(null)
  const [authPromptSecondsLeft, setAuthPromptSecondsLeft] =
    useState(AUTH_PROMPT_SECONDS)
  const { auth } = useAuthStore()
  const {
    systemName,
    logo: systemLogo,
    loading,
    logoLoaded,
  } = useSystemConfig()
  const dynamicLinks = useTopNavLinks()
  const notifications = useNotifications()
  const routerState = useRouterState()
  const pathname = routerState.location.pathname

  const user = auth.user
  const isAuthenticated = !!user
  const displaySiteName = customSiteName || systemName
  const links = dynamicLinks.length > 0 ? dynamicLinks : navLinks

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  const movePillIndicator = useCallback((element: HTMLElement | null) => {
    if (!element || !element.offsetWidth) {
      setPillIndicator(null)
      return
    }
    setPillIndicator({ left: element.offsetLeft, width: element.offsetWidth })
  }, [])

  // Park the sliding pill on the current page's link (or hide it).
  const resetPillIndicator = useCallback(() => {
    const active =
      pillRef.current?.querySelector<HTMLElement>('[data-pill-active]') ?? null
    movePillIndicator(active)
  }, [movePillIndicator])

  useEffect(() => {
    let cancelled = false
    resetPillIndicator()
    // Brand fonts may swap in after first paint and change link widths.
    document.fonts?.ready.then(() => {
      if (!cancelled) resetPillIndicator()
    })
    window.addEventListener('resize', resetPillIndicator)
    return () => {
      cancelled = true
      window.removeEventListener('resize', resetPillIndicator)
    }
  }, [resetPillIndicator, pathname, links.length, i18n.language])

  useEffect(() => {
    document.body.style.overflow = mobileOpen ? 'hidden' : ''
    return () => {
      document.body.style.overflow = ''
    }
  }, [mobileOpen])

  useEffect(() => {
    if (!authPromptTarget) return

    const intervalId = window.setInterval(() => {
      setAuthPromptSecondsLeft((seconds) => Math.max(seconds - 1, 0))
    }, 1000)

    const timeoutId = window.setTimeout(() => {
      const redirect = authPromptTarget.href
      setAuthPromptTarget(null)
      navigate({ to: '/sign-in', search: { redirect } })
    }, AUTH_PROMPT_SECONDS * 1000)

    return () => {
      window.clearInterval(intervalId)
      window.clearTimeout(timeoutId)
    }
  }, [authPromptTarget, navigate])

  const closeAuthPrompt = useCallback(() => {
    setAuthPromptTarget(null)
    setAuthPromptSecondsLeft(AUTH_PROMPT_SECONDS)
  }, [])

  const navigateToSignIn = useCallback(() => {
    const redirect = authPromptTarget?.href || '/'
    setAuthPromptTarget(null)
    navigate({ to: '/sign-in', search: { redirect } })
  }, [authPromptTarget?.href, navigate])

  const handleNavLinkClick = useCallback(
    (
      event: React.MouseEvent<HTMLAnchorElement>,
      link: TopNavLink,
      closeMobile = false
    ) => {
      if (link.disabled) {
        event.preventDefault()
        return
      }

      if (link.requiresAuth) {
        event.preventDefault()
        if (closeMobile) {
          setMobileOpen(false)
        }
        setAuthPromptSecondsLeft(AUTH_PROMPT_SECONDS)
        setAuthPromptTarget({
          title: t(link.title),
          href: link.href,
        })
        return
      }

      if (closeMobile) {
        setMobileOpen(false)
      }
    },
    [t]
  )

  return (
    <>
      <header className='pointer-events-none fixed inset-x-0 top-0 z-50'>
        {/* Backdrop fades in once the page scrolls (matches the landing nav). */}
        <div
          aria-hidden='true'
          className={cn(
            'bg-background/70 border-border absolute inset-0 border-b backdrop-blur-xl backdrop-saturate-150 transition-opacity duration-500',
            scrolled ? 'opacity-100' : 'opacity-0'
          )}
        />
        <div className='pointer-events-auto relative mx-auto w-full max-w-[1440px] px-4 md:px-[clamp(16px,4.45vw,64px)]'>
          <div
            className={cn(
              'flex items-center justify-between gap-3 transition-[height] duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] md:grid md:grid-cols-[1fr_auto_1fr]',
              scrolled ? 'h-14 md:h-16' : 'h-16 md:h-[84px]'
            )}
          >
            {/* Brand */}
            <Link
              to={homeUrl}
              className='group focus-visible:ring-ring/60 flex min-w-0 shrink-0 items-center gap-2.5 justify-self-start rounded-md outline-none focus-visible:ring-2'
            >
              {loading ? (
                <Skeleton className='size-3 rounded-none' />
              ) : (
                <PublicHeaderBrandMark
                  customLogo={customLogo}
                  systemLogo={systemLogo}
                  logoLoaded={logoLoaded}
                />
              )}
              <span className='truncate text-[17px] font-bold tracking-[-0.02em] md:text-[19px]'>
                {loading ? <Skeleton className='h-4 w-16' /> : displaySiteName}
              </span>
            </Link>

            {/* Desktop pill nav */}
            {links.length > 0 ? (
              <nav
                ref={pillRef}
                aria-label={t('Main navigation')}
                onMouseLeave={resetPillIndicator}
                onBlur={(event) => {
                  if (!event.currentTarget.contains(event.relatedTarget)) {
                    resetPillIndicator()
                  }
                }}
                className='border-border bg-card/50 font-apimix-cn relative hidden items-center gap-0.5 rounded-full border p-[5px] text-sm backdrop-blur-md md:flex'
              >
                <span
                  aria-hidden='true'
                  className='bg-foreground/[0.075] pointer-events-none absolute top-[5px] bottom-[5px] left-0 rounded-full shadow-[inset_0_0_0_1px_color-mix(in_oklch,var(--foreground)_5%,transparent)] transition-[transform,width,opacity] duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] motion-reduce:transition-none'
                  style={{
                    width: pillIndicator ? pillIndicator.width : 0,
                    transform: `translateX(${pillIndicator ? pillIndicator.left : 0}px)`,
                    opacity: pillIndicator ? 1 : 0,
                  }}
                />
                {links.map((link, i) => {
                  const isActive = isNavLinkActive(pathname, link.href)
                  const pillLinkClassName = cn(
                    'relative z-[1] rounded-full px-[18px] py-2 whitespace-nowrap transition-colors duration-300 outline-none focus-visible:ring-2 focus-visible:ring-ring/60',
                    isActive
                      ? 'text-foreground'
                      : 'text-muted-foreground hover:text-foreground',
                    link.disabled && 'pointer-events-none opacity-50'
                  )
                  const indicatorHandlers = {
                    onMouseEnter: (event: React.MouseEvent<HTMLElement>) =>
                      movePillIndicator(event.currentTarget),
                    onFocus: (event: React.FocusEvent<HTMLElement>) =>
                      movePillIndicator(event.currentTarget),
                  }
                  if (link.external) {
                    return (
                      <a
                        key={i}
                        href={link.href}
                        target='_blank'
                        rel='noopener noreferrer'
                        aria-disabled={link.disabled}
                        tabIndex={link.disabled ? -1 : undefined}
                        data-pill-active={isActive || undefined}
                        onClick={(event) => handleNavLinkClick(event, link)}
                        className={pillLinkClassName}
                        {...indicatorHandlers}
                      >
                        {t(link.title)}
                      </a>
                    )
                  }
                  return (
                    <Link
                      key={i}
                      to={link.href}
                      disabled={link.disabled}
                      aria-current={isActive ? 'page' : undefined}
                      data-pill-active={isActive || undefined}
                      onClick={(event) => handleNavLinkClick(event, link)}
                      className={pillLinkClassName}
                      {...indicatorHandlers}
                    >
                      {t(link.title)}
                    </Link>
                  )
                })}
              </nav>
            ) : (
              <span className='hidden md:block' />
            )}

            {/* Desktop utilities + primary action */}
            <div className='hidden items-center justify-end gap-1 justify-self-end md:flex'>
              {showLanguageSwitcher && <LanguageSwitcher />}
              {showThemeSwitch && <ThemeSwitch />}
              {showNotifications && (
                <NotificationPopover
                  open={notifications.popoverOpen}
                  onOpenChange={notifications.setPopoverOpen}
                  unreadCount={notifications.unreadCount}
                  activeTab={notifications.activeTab}
                  onTabChange={notifications.setActiveTab}
                  notice={notifications.notice}
                  announcements={notifications.announcements}
                  loading={notifications.loading}
                />
              )}

              {showAuthButtons && (
                <>
                  <div className='bg-border mx-1.5 h-4 w-px' />
                  {loading ? (
                    <Skeleton className='h-[42px] w-28 rounded-full' />
                  ) : isAuthenticated ? (
                    <ProfileDropdown />
                  ) : (
                    <Link
                      to='/sign-in'
                      className='brand-fill font-apimix-cn focus-visible:ring-ring/60 focus-visible:ring-offset-background inline-flex h-[42px] items-center justify-center rounded-full px-5 text-sm font-bold whitespace-nowrap transition-colors outline-none focus-visible:ring-2 focus-visible:ring-offset-2'
                    >
                      {t('Sign in / Sign up')}
                    </Link>
                  )}
                </>
              )}
            </div>

            {/* Mobile: compact actions + hamburger */}
            <div className='flex items-center gap-2 md:hidden'>
              {showThemeSwitch && <ThemeSwitch />}
              {showAuthButtons && !loading && isAuthenticated && (
                <ProfileDropdown />
              )}
              {showAuthButtons && !loading && !isAuthenticated && (
                <Link
                  to='/sign-in'
                  className='brand-fill font-apimix-cn focus-visible:ring-ring/60 inline-flex h-9 items-center justify-center rounded-full px-4 text-[13px] font-bold whitespace-nowrap outline-none focus-visible:ring-2'
                >
                  {t('Sign in')}
                </Link>
              )}
              <Button
                type='button'
                variant='ghost'
                size='icon'
                className='size-9'
                onClick={() => setMobileOpen((v) => !v)}
                aria-label={t('Toggle navigation menu')}
                aria-expanded={mobileOpen}
              >
                <div className='relative size-4'>
                  <span
                    className={cn(
                      'absolute inset-x-0 block h-[1.5px] origin-center rounded-full bg-current transition-all duration-300',
                      mobileOpen ? 'top-[7px] rotate-45' : 'top-[3px]'
                    )}
                  />
                  <span
                    className={cn(
                      'absolute inset-x-0 top-[7px] block h-[1.5px] rounded-full bg-current transition-all duration-300',
                      mobileOpen ? 'scale-x-0 opacity-0' : 'opacity-100'
                    )}
                  />
                  <span
                    className={cn(
                      'absolute inset-x-0 block h-[1.5px] origin-center rounded-full bg-current transition-all duration-300',
                      mobileOpen ? 'top-[7px] -rotate-45' : 'top-[11px]'
                    )}
                  />
                </div>
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Mobile full-screen overlay */}
      <div
        className={cn(
          'bg-background/98 fixed inset-0 z-40 backdrop-blur-2xl transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] md:pointer-events-none md:hidden',
          mobileOpen
            ? 'pointer-events-auto opacity-100'
            : 'pointer-events-none opacity-0'
        )}
      >
        <div className='flex h-full flex-col justify-between px-8 pt-20 pb-10'>
          <nav className='flex flex-col gap-1'>
            {links.map((link, i) => {
              const isActive = isNavLinkActive(pathname, link.href)
              const linkClassName = cn(
                'flex items-center gap-3 py-3 text-base font-medium tracking-tight transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)]',
                mobileOpen
                  ? 'translate-y-0 opacity-100'
                  : 'translate-y-4 opacity-0',
                isActive ? 'text-foreground' : 'text-muted-foreground',
                link.disabled && 'pointer-events-none opacity-50'
              )
              const transitionStyle = {
                transitionDelay: mobileOpen ? `${100 + i * 50}ms` : '0ms',
              }
              if (link.external) {
                return (
                  <a
                    key={i}
                    href={link.href}
                    target='_blank'
                    rel='noopener noreferrer'
                    aria-disabled={link.disabled}
                    tabIndex={link.disabled ? -1 : undefined}
                    onClick={(event) => handleNavLinkClick(event, link, true)}
                    className={linkClassName}
                    style={transitionStyle}
                  >
                    {t(link.title)}
                  </a>
                )
              }
              return (
                <Link
                  key={i}
                  to={link.href}
                  disabled={link.disabled}
                  onClick={(event) => handleNavLinkClick(event, link, true)}
                  className={linkClassName}
                  style={transitionStyle}
                >
                  {t(link.title)}
                </Link>
              )
            })}
          </nav>

          <div
            className={cn(
              'flex flex-col gap-3 transition-all duration-500',
              mobileOpen
                ? 'translate-y-0 opacity-100'
                : 'translate-y-4 opacity-0'
            )}
            style={{ transitionDelay: mobileOpen ? '250ms' : '0ms' }}
          >
            {showAuthButtons && (
              <Link
                to={isAuthenticated ? '/dashboard' : '/sign-in'}
                onClick={() => setMobileOpen(false)}
                className='brand-fill font-apimix-cn inline-flex h-11 items-center justify-center rounded-full text-sm font-bold transition-colors'
              >
                {isAuthenticated ? t('Go to Dashboard') : t('Sign in')}
              </Link>
            )}
          </div>
        </div>
      </div>

      <Dialog
        open={!!authPromptTarget}
        onOpenChange={(open) => {
          if (!open) {
            closeAuthPrompt()
          }
        }}
        title={t('Sign in required')}
        description={t('Please sign in to view {{module}}.', {
          module: authPromptTarget?.title || '',
        })}
        contentClassName='sm:max-w-md'
        contentHeight='auto'
        footer={
          <>
            <Button variant='outline' onClick={closeAuthPrompt}>
              {t('Cancel')}
            </Button>
            <Button onClick={navigateToSignIn}>{t('Sign in now')}</Button>
          </>
        }
      >
        <div className='bg-muted/40 text-muted-foreground rounded-lg px-3 py-2 text-sm'>
          {t('Redirecting to sign in in {{seconds}} seconds.', {
            seconds: authPromptSecondsLeft,
          })}
        </div>
      </Dialog>
    </>
  )
}
