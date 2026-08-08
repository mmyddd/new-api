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

import { getFreshPromptTokens } from '../format'

describe('getFreshPromptTokens', () => {
  test('returns the full prompt when no cache is reported', () => {
    assert.equal(getFreshPromptTokens(100, {}), 100)
    assert.equal(getFreshPromptTokens(100, null), 100)
    assert.equal(getFreshPromptTokens(100, undefined), 100)
  })

  test('subtracts cache read tokens', () => {
    assert.equal(
      getFreshPromptTokens(100, { cache_tokens: 30 }),
      70
    )
  })

  test('subtracts cache creation tokens when not split by duration', () => {
    assert.equal(
      getFreshPromptTokens(100, { cache_creation_tokens: 20 }),
      80
    )
  })

  test('subtracts split 5m/1h cache creation tokens', () => {
    assert.equal(
      getFreshPromptTokens(100, {
        cache_creation_tokens_5m: 10,
        cache_creation_tokens_1h: 15,
      }),
      75
    )
  })

  test('subtracts both cache read and cache write', () => {
    assert.equal(
      getFreshPromptTokens(100, {
        cache_tokens: 30,
        cache_creation_tokens: 20,
      }),
      50
    )
  })

  test('never returns a negative value (full cache hit)', () => {
    assert.equal(
      getFreshPromptTokens(50, {
        cache_tokens: 30,
        cache_creation_tokens: 30,
      }),
      0
    )
  })
})
