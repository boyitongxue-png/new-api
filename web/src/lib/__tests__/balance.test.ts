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
import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import { isLowBalanceQuota, LOW_BALANCE_THRESHOLD_USD } from '../balance'

const QUOTA_PER_USD = 500000

describe('isLowBalanceQuota', () => {
  test('flags a balance just below the USD threshold as low', () => {
    const quota = LOW_BALANCE_THRESHOLD_USD * QUOTA_PER_USD - 1
    assert.equal(isLowBalanceQuota(quota, QUOTA_PER_USD), true)
  })

  test('does not flag a balance exactly at the USD threshold', () => {
    const quota = LOW_BALANCE_THRESHOLD_USD * QUOTA_PER_USD
    assert.equal(isLowBalanceQuota(quota, QUOTA_PER_USD), false)
  })

  test('flags an empty or negative balance as low', () => {
    assert.equal(isLowBalanceQuota(0, QUOTA_PER_USD), true)
    assert.equal(isLowBalanceQuota(-100, QUOTA_PER_USD), true)
  })

  test('returns false when the quota unit size is not positive', () => {
    assert.equal(isLowBalanceQuota(0, 0), false)
    assert.equal(isLowBalanceQuota(0, -1), false)
  })

  test('returns false for a non-finite balance', () => {
    assert.equal(isLowBalanceQuota(Number.NaN, QUOTA_PER_USD), false)
  })
})
