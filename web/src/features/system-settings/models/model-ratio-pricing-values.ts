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
import { safeJsonParse } from '../utils/json-parser'

export type ModelRatioPricingValues = {
  ModelPrice: string
  ModelRatio: string
  CacheRatio: string
  CreateCacheRatio: string
  CompletionRatio: string
  ImageRatio: string
  AudioRatio: string
  AudioCompletionRatio: string
  BillingMode: string
  BillingExpr: string
}

function removeModelNames<T extends string | number>(
  value: string,
  names: string[]
) {
  const map = safeJsonParse<Record<string, T>>(value, {
    fallback: {},
    silent: true,
  })
  names.forEach((name) => delete map[name])
  return JSON.stringify(map, null, 2)
}

export function removeModelsFromPricingValues(
  values: ModelRatioPricingValues,
  names: string[]
): ModelRatioPricingValues {
  return {
    ModelPrice: removeModelNames(values.ModelPrice, names),
    ModelRatio: removeModelNames(values.ModelRatio, names),
    CacheRatio: removeModelNames(values.CacheRatio, names),
    CreateCacheRatio: removeModelNames(values.CreateCacheRatio, names),
    CompletionRatio: removeModelNames(values.CompletionRatio, names),
    ImageRatio: removeModelNames(values.ImageRatio, names),
    AudioRatio: removeModelNames(values.AudioRatio, names),
    AudioCompletionRatio: removeModelNames(values.AudioCompletionRatio, names),
    BillingMode: removeModelNames(values.BillingMode, names),
    BillingExpr: removeModelNames(values.BillingExpr, names),
  }
}
