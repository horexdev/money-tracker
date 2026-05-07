export type EvalResult = { ok: true; value: number } | { ok: false }

const MAX_LEN = 64
const MAX_VALUE = 1e10

const OPERATOR_RE = /[+\-*/×÷−()%]/

export function looksLikeExpression(value: string): boolean {
  return OPERATOR_RE.test(value)
}

export function formatForInput(value: number): string {
  return (Math.round(value * 100) / 100).toString()
}

type Term = { value: number; lonePct: boolean }

export function evaluate(input: string): EvalResult {
  if (input.length > MAX_LEN) return { ok: false }
  const src = input
    .replace(/×/g, '*')
    .replace(/÷/g, '/')
    .replace(/−/g, '-')
    .replace(/,/g, '.')
    .replace(/\s+/g, '')
  if (!src) return { ok: false }

  let pos = 0
  const peek = () => src[pos] ?? ''
  const eat = () => src[pos++] ?? ''

  const parseNumber = (): number | null => {
    const start = pos
    let dotSeen = false
    while (pos < src.length) {
      const c = src[pos]
      if (c >= '0' && c <= '9') { pos++; continue }
      if (c === '.') {
        if (dotSeen) return null
        dotSeen = true
        pos++
        continue
      }
      break
    }
    if (start === pos) return null
    const n = parseFloat(src.slice(start, pos))
    return isNaN(n) ? null : n
  }

  const parsePrimary = (): number | null => {
    if (peek() === '(') {
      pos++
      const v = parseAdd()
      if (v === null) return null
      if (peek() !== ')') return null
      pos++
      return v
    }
    return parseNumber()
  }

  const parsePostfix = (): Term | null => {
    const v = parsePrimary()
    if (v === null) return null
    if (peek() === '%') {
      pos++
      return { value: v / 100, lonePct: true }
    }
    return { value: v, lonePct: false }
  }

  const parseUnary = (): Term | null => {
    if (peek() === '-') {
      pos++
      const t = parseUnary()
      if (t === null) return null
      return { value: -t.value, lonePct: t.lonePct }
    }
    if (peek() === '+') {
      pos++
      return parseUnary()
    }
    return parsePostfix()
  }

  const parseMul = (): Term | null => {
    const left = parseUnary()
    if (left === null) return null
    if (peek() !== '*' && peek() !== '/') return left
    let value = left.value
    while (peek() === '*' || peek() === '/') {
      const op = eat()
      const right = parseUnary()
      if (right === null) return null
      if (op === '*') value *= right.value
      else {
        if (right.value === 0) return null
        value /= right.value
      }
    }
    return { value, lonePct: false }
  }

  const parseAdd = (): number | null => {
    const leftT = parseMul()
    if (leftT === null) return null
    let acc = leftT.value
    while (peek() === '+' || peek() === '-') {
      const op = eat()
      const right = parseMul()
      if (right === null) return null
      const delta = right.lonePct ? acc * right.value : right.value
      acc = op === '+' ? acc + delta : acc - delta
    }
    return acc
  }

  const result = parseAdd()
  if (result === null) return { ok: false }
  if (pos !== src.length) return { ok: false }
  if (!isFinite(result)) return { ok: false }
  if (Math.abs(result) > MAX_VALUE) return { ok: false }
  return { ok: true, value: result }
}
