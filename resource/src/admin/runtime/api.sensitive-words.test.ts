import { afterEach, describe, expect, it, vi } from 'vitest'
import { getSensitiveWordSettings, getSensitiveWords, saveSensitiveWord } from './api'

function respond(result: unknown) {
  vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ code: 0, result }), {
    headers: { 'Content-Type': 'application/json' },
  })))
}

afterEach(() => vi.unstubAllGlobals())

describe('sensitive word API response contract', () => {
  it('unwraps settings so the saved enabled state is displayed', async () => {
    const settings = { enabled: true, mode: 'after_review' }
    respond({ settings })
    await expect(getSensitiveWordSettings()).resolves.toEqual(settings)
  })
  it('unwraps the dictionary for rendering and filtering', async () => {
    const words = [{ id: 1, word: 'example', action: 'reject', enabled: true, replacement: '' }]
    respond({ words })
    await expect(getSensitiveWords()).resolves.toEqual(words)
  })
  it.each([null, []])('normalizes an empty dictionary: %s', async (words) => {
    respond({ words })
    await expect(getSensitiveWords()).resolves.toEqual([])
  })
  it('reports a malformed payload before rendering', async () => {
    respond({ words: {} })
    await expect(getSensitiveWords()).rejects.toThrow()
  })
  it('keeps the saved word envelope used by the editor', async () => {
    const word = { id: 1, word: 'example', action: 'reject' as const, enabled: true, replacement: '' }
    respond({ word })
    await expect(saveSensitiveWord(word)).resolves.toEqual({ word })
  })
})
