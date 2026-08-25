import { describe, expect, it } from 'vitest'
import { imageFilenameForMime } from '../src/runtime/image'

describe('image processing filenames', () => {
  it('keeps the filename extension consistent with converted MIME type', () => {
    expect(imageFilenameForMime('photo.png', 'image/webp')).toBe('photo.webp')
    expect(imageFilenameForMime('photo', 'image/webp')).toBe('photo.webp')
    expect(imageFilenameForMime('archive.photo.webp', 'image/jpeg')).toBe('archive.photo.jpg')
  })
})
