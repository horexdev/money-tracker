import { describe, expect, it } from 'vitest'

import { evaluate, formatForInput, looksLikeExpression } from './expression'

describe('looksLikeExpression', () => {
  it.each([
    ['1+2', true],
    ['10*5', true],
    ['100-20', true],
    ['10×5', true],
    ['10÷5', true],
    ['200%', true],
    ['(1+2)', true],
    ['100', false],
    ['1500.50', false],
    ['', false],
  ])('looksLikeExpression(%j) → %s', (input, expected) => {
    expect(looksLikeExpression(input)).toBe(expected)
  })
})

describe('evaluate — happy path', () => {
  it('adds integers', () => {
    expect(evaluate('1+2')).toEqual({ ok: true, value: 3 })
  })

  it('multiplies integers', () => {
    expect(evaluate('2*3')).toEqual({ ok: true, value: 6 })
  })

  it('handles decimal numbers', () => {
    expect(evaluate('1.5+2.5')).toEqual({ ok: true, value: 4 })
  })

  it('respects operator precedence (* before +)', () => {
    expect(evaluate('2+3*4')).toEqual({ ok: true, value: 14 })
  })

  it('honours parentheses', () => {
    expect(evaluate('(2+3)*4')).toEqual({ ok: true, value: 20 })
  })

  it('handles unary minus', () => {
    expect(evaluate('-5+10')).toEqual({ ok: true, value: 5 })
  })

  it('replaces unicode × and ÷ with * and /', () => {
    expect(evaluate('2×3÷6')).toEqual({ ok: true, value: 1 })
  })

  it('replaces unicode minus with ASCII', () => {
    expect(evaluate('10−5')).toEqual({ ok: true, value: 5 })
  })

  it('treats comma as a decimal separator', () => {
    expect(evaluate('1,5+0,5')).toEqual({ ok: true, value: 2 })
  })

  it('ignores whitespace', () => {
    expect(evaluate(' 1 + 2 ')).toEqual({ ok: true, value: 3 })
  })
})

describe('evaluate — percentage', () => {
  it('treats a lone percent as a fraction', () => {
    expect(evaluate('10%')).toEqual({ ok: true, value: 0.1 })
  })

  it('applies percentage of accumulator on +', () => {
    expect(evaluate('200+10%')).toEqual({ ok: true, value: 220 })
  })

  it('applies percentage of accumulator on -', () => {
    expect(evaluate('200-10%')).toEqual({ ok: true, value: 180 })
  })
})

describe('evaluate — failure modes', () => {
  it('rejects empty input', () => {
    expect(evaluate('')).toEqual({ ok: false })
  })

  it('rejects input over MAX_LEN (64) chars', () => {
    const long = '1+'.repeat(40) + '1'
    expect(evaluate(long).ok).toBe(false)
  })

  it('rejects an operator without a right operand', () => {
    expect(evaluate('1+')).toEqual({ ok: false })
    expect(evaluate('1*2*')).toEqual({ ok: false })
  })

  it('rejects unbalanced parentheses', () => {
    expect(evaluate('(1+2')).toEqual({ ok: false })
  })

  it('rejects a bare operator', () => {
    expect(evaluate('+')).toEqual({ ok: false })
  })

  it('rejects two decimal points in one number', () => {
    expect(evaluate('1.2.3')).toEqual({ ok: false })
  })

  it('rejects division by zero', () => {
    expect(evaluate('1/0')).toEqual({ ok: false })
  })

  it('rejects results above MAX_VALUE (1e10)', () => {
    expect(evaluate('99999999999')).toEqual({ ok: false })
  })
})

describe('formatForInput', () => {
  it('rounds to two decimals', () => {
    expect(formatForInput(1.234)).toBe('1.23')
    expect(formatForInput(1.236)).toBe('1.24')
  })

  it('drops trailing zeros via toString', () => {
    expect(formatForInput(1.5)).toBe('1.5')
    expect(formatForInput(2)).toBe('2')
  })
})
