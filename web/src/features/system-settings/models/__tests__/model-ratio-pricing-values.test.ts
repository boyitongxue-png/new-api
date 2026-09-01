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
import { describe, expect, test } from 'vitest'

import {
  removeModelsFromPricingValues,
  type ModelRatioPricingValues,
} from '../model-ratio-pricing-values'

const fields: Array<keyof ModelRatioPricingValues> = [
  'ModelPrice',
  'ModelRatio',
  'CacheRatio',
  'CreateCacheRatio',
  'CompletionRatio',
  'ImageRatio',
  'AudioRatio',
  'AudioCompletionRatio',
  'BillingMode',
  'BillingExpr',
]

test('removes selected models from every persisted pricing map', () => {
  const values = Object.fromEntries(
    fields.map((field) => [
      field,
      JSON.stringify({ 'remove-a': 1, 'remove-b': 2, retain: 3 }),
    ])
  ) as ModelRatioPricingValues

  const result = removeModelsFromPricingValues(values, ['remove-a', 'remove-b'])

  fields.forEach((field) => {
    expect(JSON.parse(result[field])).toEqual({ retain: 3 })
  })
})

describe('removeModelsFromPricingValues', () => {
  test('keeps unaffected model values', () => {
    const values: ModelRatioPricingValues = {
      ModelPrice: '{"remove":0.1,"retain":0.2}',
      ModelRatio: '{"remove":1,"retain":2}',
      CacheRatio: '{"remove":0.1,"retain":0.2}',
      CreateCacheRatio: '{"remove":0.1,"retain":0.2}',
      CompletionRatio: '{"remove":0.1,"retain":0.2}',
      ImageRatio: '{"remove":0.1,"retain":0.2}',
      AudioRatio: '{"remove":0.1,"retain":0.2}',
      AudioCompletionRatio: '{"remove":0.1,"retain":0.2}',
      BillingMode: '{"remove":"per-request","retain":"tiered_expr"}',
      BillingExpr: '{"remove":"1","retain":"2"}',
    }

    const result = removeModelsFromPricingValues(values, ['remove'])

    expect(JSON.parse(result.ModelPrice)).toEqual({ retain: 0.2 })
    expect(JSON.parse(result.BillingMode)).toEqual({ retain: 'tiered_expr' })
    expect(JSON.parse(result.BillingExpr)).toEqual({ retain: '2' })
  })
})
