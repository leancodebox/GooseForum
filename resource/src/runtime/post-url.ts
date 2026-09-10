// Include the post window in the path so a permalink also works on a fresh visit.
export function postURL(topicId: number, postNo = 1, postId?: number) {
  const path = `/p/post/${topicId}${postNo > 1 ? `/${postNo}` : ''}`
  return postId ? `${path}#post-${postId}` : path
}
