export interface PostViewportRect {
  postNo: number
  top: number
  bottom: number
  height: number
}

export interface PostViewportProgressInput {
  posts: PostViewportRect[]
  markerY: number
  viewportTop: number
  viewportBottom: number
  maxPostNo: number
  visibleSlotSize: number
}

export interface PostViewportProgress {
  postNo: number
  current: number
  start: number
  end: number
}

export function measurePostViewportProgressFromRects(input: PostViewportProgressInput): PostViewportProgress {
  const fallbackPostNo = fallbackPostNoFromRects(input.posts, input.viewportTop, input.viewportBottom)
  let coveringPostNo: number | null = null
  let coveringDistance = Number.POSITIVE_INFINITY
  let nearestPostNo: number | null = null
  let nearestDistance = Number.POSITIVE_INFINITY

  for (const post of input.posts) {
    if (!post.postNo) continue
    if (post.bottom <= input.viewportTop || post.top >= input.viewportBottom) continue

    const visibleTop = Math.max(input.viewportTop, post.top)
    const visibleBottom = Math.min(input.viewportBottom, post.bottom)
    if (visibleBottom <= visibleTop) continue

    if (post.top <= input.markerY && post.bottom >= input.markerY) {
      const distance = Math.abs(post.top - input.markerY)
      if (distance < coveringDistance) {
        coveringPostNo = post.postNo
        coveringDistance = distance
      }
      continue
    }

    const distance = Math.abs(post.top - input.markerY)
    if (distance < nearestDistance) {
      nearestPostNo = post.postNo
      nearestDistance = distance
    }
  }

  const postNo = coveringPostNo ?? nearestPostNo ?? fallbackPostNo
  const current = progressForPostNo(input.maxPostNo, postNo)

  return {
    postNo,
    current,
    start: Math.max(0, current - input.visibleSlotSize / 2),
    end: Math.min(1, current + input.visibleSlotSize / 2),
  }
}

function fallbackPostNoFromRects(posts: PostViewportRect[], viewportTop: number, viewportBottom: number) {
  const validPosts = posts.filter((post) => post.postNo > 0)
  if (!validPosts.length) return 1
  if (validPosts.every((post) => post.bottom <= viewportTop)) return validPosts[validPosts.length - 1].postNo
  if (validPosts.every((post) => post.top >= viewportBottom)) return validPosts[0].postNo
  return validPosts[0].postNo
}

function progressForPostNo(maxPostNo: number, postNo: number) {
  if (maxPostNo <= 1) return 0
  return Math.min(1, Math.max(0, (Math.max(1, postNo) - 1) / (maxPostNo - 1)))
}
