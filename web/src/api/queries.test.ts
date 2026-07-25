import { describe, expect, it } from 'vitest'

import { publicQueryKeys } from './queries'

describe('public query options', () => {
  it('uses the same discovery key for equivalent filters', () => {
    const first = publicQueryKeys.discovery({
      campus_id: '550e8400-e29b-41d4-a716-446655440002',
      category: 'single_room,self_contained',
      area: ' South Gate ',
      page: 1,
    })
    const second = publicQueryKeys.discovery({
      campus_id: '550e8400-e29b-41d4-a716-446655440002',
      category: 'self_contained, single_room,single_room',
      area: 'South Gate',
      page: 1,
    })

    expect(first).toEqual(second)
  })
})
