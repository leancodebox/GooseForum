import { describe, expect, test } from 'vitest'
import { measurePostViewportProgressFromRects } from '../src/runtime/post-viewport-progress'

describe('measurePostViewportProgressFromRects', () => {
  test('falls back to the last loaded post when all posts are above the viewport', () => {
    const progress = measurePostViewportProgressFromRects({
      posts: [
        { postNo: 1, top: -160, bottom: -80, height: 80 },
        { postNo: 2, top: -70, bottom: -20, height: 50 },
      ],
      markerY: 300,
      viewportTop: 88,
      viewportBottom: 800,
      maxPostNo: 5,
      visibleSlotSize: 0.2,
    })

    expect(progress.postNo).toBe(2)
    expect(progress.current).toBe(0.25)
  })

  test('falls back to the first loaded post when all posts are below the viewport', () => {
    const progress = measurePostViewportProgressFromRects({
      posts: [
        { postNo: 3, top: 840, bottom: 920, height: 80 },
        { postNo: 4, top: 930, bottom: 1010, height: 80 },
      ],
      markerY: 300,
      viewportTop: 88,
      viewportBottom: 800,
      maxPostNo: 5,
      visibleSlotSize: 0.2,
    })

    expect(progress.postNo).toBe(3)
  })

  test('uses the post covering the viewport marker when one is visible', () => {
    const progress = measurePostViewportProgressFromRects({
      posts: [
        { postNo: 1, top: 120, bottom: 220, height: 100 },
        { postNo: 2, top: 240, bottom: 380, height: 140 },
      ],
      markerY: 300,
      viewportTop: 88,
      viewportBottom: 800,
      maxPostNo: 5,
      visibleSlotSize: 0.2,
    })

    expect(progress.postNo).toBe(2)
    expect(progress.current).toBe(0.25)
  })

  test('keeps the first post anchored at the start of the rail', () => {
    const progress = measurePostViewportProgressFromRects({
      posts: [
        { postNo: 1, top: -80, bottom: 260, height: 340 },
        { postNo: 2, top: 280, bottom: 360, height: 80 },
      ],
      markerY: 120,
      viewportTop: 88,
      viewportBottom: 800,
      maxPostNo: 5,
      visibleSlotSize: 0.2,
    })

    expect(progress.postNo).toBe(1)
    expect(progress.current).toBe(0)
  })
})
