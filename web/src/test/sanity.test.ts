import { describe, expect, it } from 'vitest'

describe('test runner sanity', () => {
  it('runs and asserts truthy results', () => {
    expect(1 + 1).toBe(2)
    expect('hello').toMatch(/ell/)
  })
})
