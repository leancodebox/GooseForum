import React from 'react'
import { type Post } from '../data/schema'

type PostActionType = 'view' | 'approve' | 'reject' | 'top' | 'recommend'

interface PostsContextType {
  open: PostActionType | null
  setOpen: (open: PostActionType | null) => void
  currentRow: Post | null
  setCurrentRow: (row: Post | null) => void
}

const PostsContext = React.createContext<PostsContextType | null>(null)

interface Props {
  children: React.ReactNode
}

export function PostsProvider({ children }: Props) {
  const [open, setOpen] = React.useState<PostActionType | null>(null)
  const [currentRow, setCurrentRow] = React.useState<Post | null>(null)

  return (
    <PostsContext.Provider value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </PostsContext.Provider>
  )
}

export const usePosts = () => {
  const context = React.useContext(PostsContext)
  if (!context) {
    throw new Error('usePosts must be used within a PostsProvider')
  }
  return context
}
