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
/**
 * Account balance helpers shared by the console header balance chip.
 */

/** Balances below this many USD are surfaced with a low-balance warning. */
export const LOW_BALANCE_THRESHOLD_USD = 5

/** React Query key of the header's fresh `/api/user/self` balance read. */
export const HEADER_BALANCE_QUERY_KEY = ['header', 'self-balance'] as const

/**
 * Whether a quota balance (in quota units, where `quotaPerUnit` units equal
 * 1 USD) is below {@link LOW_BALANCE_THRESHOLD_USD}. Returns `false` when the
 * inputs cannot be compared (non-finite quota or a non-positive unit size).
 */
export function isLowBalanceQuota(quota: number, quotaPerUnit: number): boolean {
  if (!Number.isFinite(quota) || !Number.isFinite(quotaPerUnit)) return false
  if (quotaPerUnit <= 0) return false
  return quota < LOW_BALANCE_THRESHOLD_USD * quotaPerUnit
}
