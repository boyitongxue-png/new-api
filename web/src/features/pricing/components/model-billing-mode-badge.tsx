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

import { cn } from '@/lib/utils'

import { isDynamicPricingModel } from '../lib/dynamic-price'
import { isTokenBasedModel } from '../lib/model-helpers'
import {
  hasImageResolutionPricing,
  hasTaskResolutionPricing,
} from '../lib/price'
import type { PricingModel } from '../types'

interface ModelBillingModeBadgeProps {
  model: PricingModel
  className?: string
}

/**
 * Restrained pill treatments, one per billing family:
 * - `outline`: metered by tokens (the common case)
 * - `dashed`: flat per request / per second
 * - `filled`: pricing that varies with the request (dynamic, resolution)
 */
type BillingTagStyle = 'outline' | 'dashed' | 'filled'

const BILLING_TAG_STYLE_CLASSES: Record<BillingTagStyle, string> = {
  outline: 'border-border text-foreground/85',
  dashed: 'border-dashed border-foreground/30 text-muted-foreground',
  filled: 'border-transparent bg-primary/15 text-foreground',
}

export function ModelBillingModeBadge(props: ModelBillingModeBadgeProps) {
  const { t } = useTranslation()
  let label = t('Per Request')
  let tagStyle: BillingTagStyle = 'dashed'

  if (hasTaskResolutionPricing(props.model)) {
    label = t('Video resolution pricing')
    tagStyle = 'filled'
  } else if (hasImageResolutionPricing(props.model)) {
    label = t('Image resolution pricing')
    tagStyle = 'filled'
  } else if (props.model.billing_mode === 'per-second') {
    label = t('Per second')
    tagStyle = 'dashed'
  } else if (isDynamicPricingModel(props.model)) {
    label = t('Dynamic Pricing')
    tagStyle = 'filled'
  } else if (isTokenBasedModel(props.model)) {
    label = t('Token-based')
    tagStyle = 'outline'
  }

  return (
    <span
      className={cn(
        'inline-flex h-5 items-center rounded-full border px-2 text-[11px] leading-none font-medium whitespace-nowrap',
        BILLING_TAG_STYLE_CLASSES[tagStyle],
        props.className
      )}
    >
      {label}
    </span>
  )
}
