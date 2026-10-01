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
import { Add01Icon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { useIsSidebarModuleVisible } from '@/hooks/use-sidebar-config'
import { useSystemConfig } from '@/hooks/use-system-config'
import { getSelf } from '@/lib/api'
import { HEADER_BALANCE_QUERY_KEY, isLowBalanceQuota } from '@/lib/balance'
import { formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

type SelfBalanceResponse = {
  success?: boolean
  data?: { quota?: number }
}

/**
 * Console top-bar balance chip + lime top-up button (links to /wallet).
 *
 * The balance is re-read from `/api/user/self` (cached for a minute and on
 * window focus) so it reflects spending since sign-in, falling back to the
 * signed-in user snapshot. Below the low-balance threshold the chip turns
 * amber. On narrow screens only the top-up button is shown.
 */
export function HeaderBalance() {
  const { t } = useTranslation()
  const authUser = useAuthStore((state) => state.auth.user)
  const { currency } = useSystemConfig()
  const isWalletVisible = useIsSidebarModuleVisible('/wallet')
  const isSignedIn = Boolean(authUser)

  const selfQuery = useQuery({
    queryKey: HEADER_BALANCE_QUERY_KEY,
    queryFn: async () => (await getSelf()) as SelfBalanceResponse,
    enabled: isSignedIn && isWalletVisible,
    staleTime: 60 * 1000,
    refetchOnWindowFocus: true,
  })

  if (!authUser || !isWalletVisible) return null

  const freshQuota = selfQuery.data?.success
    ? selfQuery.data.data?.quota
    : undefined
  const quota = Number(freshQuota ?? authUser.quota ?? 0)
  const isLow = isLowBalanceQuota(quota, currency.quotaPerUnit)
  const balanceText = formatQuota(quota)

  return (
    <div className='flex items-center gap-1.5'>
      <Link
        to='/wallet'
        aria-label={
          isLow
            ? t('Low balance: {{balance}}. Open wallet', {
                balance: balanceText,
              })
            : t('Balance: {{balance}}. Open wallet', { balance: balanceText })
        }
        className={cn(
          'focus-visible:ring-ring/50 hidden h-8 items-center gap-2 rounded-full border px-3 text-xs whitespace-nowrap transition-colors outline-none focus-visible:ring-2 sm:inline-flex',
          isLow
            ? 'border-[var(--brand-warn-line,color-mix(in_oklch,var(--warning)_45%,transparent))] bg-[var(--brand-warn-bg,color-mix(in_oklch,var(--warning)_10%,transparent))] text-[var(--brand-warn,var(--warning))]'
            : 'border-border text-muted-foreground hover:text-foreground hover:bg-muted/60'
        )}
      >
        <span
          aria-hidden='true'
          className={cn(
            'size-1.5 shrink-0 rounded-full',
            isLow ? 'bg-current' : 'bg-[var(--brand-lime,var(--primary))]'
          )}
        />
        <span>{isLow ? t('Low balance') : t('Balance')}</span>
        <span
          className={cn(
            'font-mono font-semibold tabular-nums',
            !isLow && 'text-foreground'
          )}
        >
          {balanceText}
        </span>
      </Link>
      <Link
        to='/wallet'
        className='brand-fill focus-visible:ring-ring/60 focus-visible:ring-offset-background inline-flex h-8 items-center gap-1 rounded-full px-3 text-xs font-bold whitespace-nowrap transition-colors outline-none focus-visible:ring-2 focus-visible:ring-offset-2'
      >
        <HugeiconsIcon
          icon={Add01Icon}
          strokeWidth={2.4}
          className='size-3.5'
          aria-hidden='true'
        />
        {t('Top up')}
      </Link>
    </div>
  )
}
