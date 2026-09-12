// Include the post window in the path so a permalink also works on a fresh visit.
export function postURL(topicId: number, postNo = 1) {
  return `/p/post/${topicId}${postNo > 1 ? `/${postNo}` : ''}`
}
