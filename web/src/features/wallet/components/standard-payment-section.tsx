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
import { ArrowRight, CheckCircle2, Loader2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'

import { formatCurrency, getMinTopupAmount, getPaymentIcon } from '../lib'
import type { PaymentMethod, TopupInfo } from '../types'

interface StandardPaymentSectionProps {
  topupInfo: TopupInfo | null
  topupAmount: number
  paymentAmount: number
  calculating: boolean
  paymentLoading: string | null
  onPay: (method: PaymentMethod) => void
}

/**
 * Payment method picker + order summary + primary pay button.
 *
 * Picking a method only selects it (lime/primary selected state); the large
 * pay button then hands the chosen method to the existing
 * `onPaymentMethodSelect` flow, which validates, recalculates and opens the
 * confirmation dialog exactly as before.
 */
export function StandardPaymentSection(props: StandardPaymentSectionProps) {
  const { t } = useTranslation()
  const [chosenType, setChosenType] = useState<string | null>(null)

  const baseMinTopup = getMinTopupAmount(props.topupInfo)
  const options = (props.topupInfo?.pay_methods ?? []).map((method) => {
    const minTopup = Math.max(method.min_topup || 0, baseMinTopup)
    return { method, minTopup, disabled: minTopup > props.topupAmount }
  })
  const chosen =
    options.find(
      (option) => option.method.type === chosenType && !option.disabled
    ) ??
    options.find((option) => !option.disabled) ??
    null
  const isPaying =
    chosen !== null && props.paymentLoading === chosen.method.type

  return (
    <div className='space-y-4 sm:space-y-5'>
      <div className='space-y-2.5 sm:space-y-3'>
        <Label
          id='wallet-payment-method-label'
          className='text-muted-foreground text-xs font-medium tracking-wider uppercase'
        >
          {t('Payment Method')}
        </Label>
        <div
          role='group'
          aria-labelledby='wallet-payment-method-label'
          className='grid grid-cols-1 gap-2 min-[420px]:grid-cols-2 sm:gap-3 lg:grid-cols-3'
        >
          {options.map((option) => {
            const selected = chosen?.method.type === option.method.type
            const disabledReason = option.disabled
              ? t('Minimum topup amount: {{amount}}', {
                  amount: option.minTopup,
                })
              : undefined
            return (
              <Button
                key={option.method.type}
                type='button'
                variant='outline'
                aria-pressed={selected}
                disabled={option.disabled || !!props.paymentLoading}
                title={disabledReason}
                aria-label={
                  disabledReason
                    ? `${option.method.name}. ${disabledReason}`
                    : option.method.name
                }
                onClick={() => setChosenType(option.method.type)}
                className={cn(
                  'h-auto min-h-14 min-w-0 justify-start gap-3 rounded-xl px-3.5 py-2.5 text-left',
                  selected &&
                    'border-primary bg-primary/10 ring-primary hover:bg-primary/15 dark:border-primary dark:bg-primary/10 dark:hover:bg-primary/15 ring-1'
                )}
              >
                <span className='bg-muted flex size-8 shrink-0 items-center justify-center rounded-lg'>
                  {getPaymentIcon(
                    option.method.type,
                    'h-4 w-4',
                    option.method.icon,
                    option.method.name
                  )}
                </span>
                <span className='flex min-w-0 flex-1 flex-col items-start gap-0.5'>
                  <span className='max-w-full truncate font-semibold'>
                    {option.method.name}
                  </span>
                  {option.disabled && (
                    <span className='text-muted-foreground max-w-full truncate text-[11px] leading-4 font-normal'>
                      {`${t('Minimum:')} ${option.minTopup}`}
                    </span>
                  )}
                </span>
                {selected ? (
                  <CheckCircle2
                    className='text-primary size-5 shrink-0'
                    aria-hidden='true'
                  />
                ) : (
                  <span
                    aria-hidden='true'
                    className='border-border size-5 shrink-0 rounded-full border'
                  />
                )}
              </Button>
            )
          })}
        </div>
      </div>

      <div className='bg-muted/40 flex flex-col gap-3 rounded-xl border p-3 sm:flex-row sm:items-center sm:justify-between sm:p-4'>
        <div className='flex items-baseline justify-between gap-3 sm:block'>
          <div className='text-muted-foreground text-xs'>
            {t('Amount to pay:')}
          </div>
          {props.calculating ? (
            <Skeleton className='mt-1 h-8 w-28' />
          ) : (
            <div className='font-mono text-2xl font-bold tracking-tight tabular-nums sm:text-3xl'>
              {formatCurrency(props.paymentAmount)}
            </div>
          )}
        </div>
        <Button
          type='button'
          size='lg'
          disabled={!chosen || !!props.paymentLoading}
          onClick={() => {
            if (chosen) props.onPay(chosen.method)
          }}
          className='h-12 w-full gap-2 rounded-full px-8 text-base font-bold sm:w-auto sm:min-w-56'
        >
          {isPaying ? (
            <Loader2 className='size-4 animate-spin' aria-hidden='true' />
          ) : null}
          {t('Pay now')}
          {!isPaying ? <ArrowRight className='size-4' aria-hidden='true' /> : null}
        </Button>
      </div>
    </div>
  )
}
