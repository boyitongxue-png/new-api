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
import { describe, expect, it } from 'vitest'

import {
  buildPricingOptionUpdates,
  emptyModelCost,
  resolveModelCostEntry,
} from '../model-commercial'

describe('model commercial pricing updates', () => {
  it('switches a token-priced model to per-second pricing without changing other models', () => {
    const updates = buildPricingOptionUpdates(
      {
        ModelPrice: JSON.stringify({ other: 0.2 }),
        ModelRatio: JSON.stringify({ video: 1, other: 2 }),
        CompletionRatio: JSON.stringify({ video: 4 }),
        'billing_setting.billing_mode': '{}',
      },
      {
        name: 'video',
        billingMode: 'per-second',
        price: '0.08',
        taskBillingPricing: JSON.stringify({
          mode: 'per-second',
          resolution_prices: { '1080p': 0.12 },
        }),
      }
    )

    const values = Object.fromEntries(
      updates.map((update) => [update.key, JSON.parse(update.value)])
    )
    expect(values.ModelPrice).toEqual({ other: 0.2, video: 0.08 })
    expect(values.ModelRatio).toEqual({ other: 2 })
    expect(values.CompletionRatio).toEqual({})
    expect(values['billing_setting.billing_mode']).toEqual({
      video: 'per-second',
    })
    expect(values['billing_setting.task_billing_pricing']).toEqual({
      video: {
        mode: 'per-second',
        resolution_prices: { '1080p': 0.12 },
      },
    })
  })

  it('resolves channel cost before model and default fallbacks', () => {
    const modelCost = emptyModelCost()
    modelCost.video_per_second = 0.8
    const channelCost = emptyModelCost()
    channelCost.video_per_second = 1.2
    const fallbackCost = emptyModelCost()
    fallbackCost.video_per_second = 0.5

    expect(
      resolveModelCostEntry(
        {
          default: fallbackCost,
          video: modelCost,
          'channel:7:video': channelCost,
        },
        'video',
        7
      )
    ).toEqual(channelCost)
    expect(resolveModelCostEntry({ video: modelCost }, 'video', 8)).toEqual(
      modelCost
    )
    expect(
      resolveModelCostEntry({ default: fallbackCost }, 'video', 8)
    ).toEqual(fallbackCost)
  })
})
