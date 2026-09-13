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
import { combineBillingExpr } from '@/features/pricing/lib/billing-expr'
import type { ModelRatioData } from '@/features/system-settings/models/model-pricing-sheet'
import { safeJsonParse } from '@/features/system-settings/utils/json-parser'

export type ModelCostEntry = {
  currency: string
  enabled: boolean
  input_per_1m: number
  output_per_1m: number
  cache_read_per_1m: number
  cache_write_per_1m: number
  image_token_per_1m: number
  image_per_unit: number
  audio_input_per_1m: number
  audio_output_per_1m: number
  audio_input_per_second: number
  audio_output_per_second: number
  video_per_second: number
  video_per_second_by_resolution: Record<string, number>
  request_fee: number
}

export type ModelCostMap = Record<string, ModelCostEntry>

export function resolveModelCostEntry(
  costs: ModelCostMap,
  modelName: string,
  channelId?: number
): ModelCostEntry | undefined {
  const keys = [
    channelId !== undefined ? `channel:${channelId}:${modelName}` : '',
    modelName,
    'default',
  ].filter(Boolean)
  for (const key of keys) {
    const cost = costs[key]
    if (cost?.enabled) return cost
  }
  return undefined
}

export const emptyModelCost = (): ModelCostEntry => ({
  currency: 'USD',
  enabled: true,
  input_per_1m: 0,
  output_per_1m: 0,
  cache_read_per_1m: 0,
  cache_write_per_1m: 0,
  image_token_per_1m: 0,
  image_per_unit: 0,
  audio_input_per_1m: 0,
  audio_output_per_1m: 0,
  audio_input_per_second: 0,
  audio_output_per_second: 0,
  video_per_second: 0,
  video_per_second_by_resolution: {},
  request_fee: 0,
})

export function parseModelCostMap(raw: string): ModelCostMap {
  const parsed = safeJsonParse<Record<string, Partial<ModelCostEntry>>>(raw, {
    fallback: {},
    silent: true,
  })
  return Object.fromEntries(
    Object.entries(parsed).map(([key, value]) => [
      key,
      { ...emptyModelCost(), ...value },
    ])
  )
}

export function getOptionMap(
  options: Array<{ key: string; value: string }> | undefined
): Record<string, string> {
  return Object.fromEntries(
    (options || []).map((option) => [option.key, option.value])
  )
}

export function buildPricingOptionUpdates(
  options: Record<string, string>,
  data: ModelRatioData
): Array<{ key: string; value: string }> {
  const numberKeys = [
    'ModelPrice',
    'ModelRatio',
    'CacheRatio',
    'CreateCacheRatio',
    'CompletionRatio',
    'ImageRatio',
    'AudioRatio',
    'AudioCompletionRatio',
  ] as const
  const numberMaps = Object.fromEntries(
    numberKeys.map((key) => [
      key,
      safeJsonParse<Record<string, number>>(options[key] || '{}', {
        fallback: {},
        silent: true,
      }),
    ])
  ) as Record<(typeof numberKeys)[number], Record<string, number>>
  const modeMap = safeJsonParse<Record<string, string>>(
    options['billing_setting.billing_mode'] || '{}',
    { fallback: {}, silent: true }
  )
  const expressionMap = safeJsonParse<Record<string, string>>(
    options['billing_setting.billing_expr'] || '{}',
    { fallback: {}, silent: true }
  )
  const taskMap = safeJsonParse<Record<string, unknown>>(
    options['billing_setting.task_billing_pricing'] || '{}',
    { fallback: {}, silent: true }
  )
  const discountMap = safeJsonParse<Record<string, unknown>>(
    options['billing_setting.scheduled_discount'] || '{}',
    { fallback: {}, silent: true }
  )

  numberKeys.forEach((key) => delete numberMaps[key][data.name])
  delete modeMap[data.name]
  delete expressionMap[data.name]
  delete taskMap[data.name]
  delete discountMap[data.name]

  const setNumber = (
    key: (typeof numberKeys)[number],
    value: string | undefined
  ) => {
    if (!value) return
    const parsed = Number(value)
    if (Number.isFinite(parsed)) numberMaps[key][data.name] = parsed
  }

  if (data.billingMode === 'tiered_expr') {
    const expression = combineBillingExpr(
      data.billingExpr || '',
      data.requestRuleExpr || ''
    )
    modeMap[data.name] = 'tiered_expr'
    if (expression) expressionMap[data.name] = expression
  } else if (
    data.billingMode === 'per-request' ||
    data.billingMode === 'per-second'
  ) {
    modeMap[data.name] = data.billingMode
    setNumber('ModelPrice', data.price)
  } else {
    setNumber('ModelRatio', data.ratio)
    setNumber('CacheRatio', data.cacheRatio)
    setNumber('CreateCacheRatio', data.createCacheRatio)
    setNumber('CompletionRatio', data.completionRatio)
    setNumber('ImageRatio', data.imageRatio)
    setNumber('AudioRatio', data.audioRatio)
    setNumber('AudioCompletionRatio', data.audioCompletionRatio)
  }

  if (data.taskBillingPricing) {
    taskMap[data.name] = JSON.parse(data.taskBillingPricing) as unknown
  }
  if (data.scheduledDiscount) {
    discountMap[data.name] = JSON.parse(data.scheduledDiscount) as unknown
  }

  return [
    ...numberKeys.map((key) => ({
      key,
      value: JSON.stringify(numberMaps[key], null, 2),
    })),
    {
      key: 'billing_setting.billing_mode',
      value: JSON.stringify(modeMap, null, 2),
    },
    {
      key: 'billing_setting.billing_expr',
      value: JSON.stringify(expressionMap, null, 2),
    },
    {
      key: 'billing_setting.task_billing_pricing',
      value: JSON.stringify(taskMap, null, 2),
    },
    {
      key: 'billing_setting.scheduled_discount',
      value: JSON.stringify(discountMap, null, 2),
    },
  ]
}
