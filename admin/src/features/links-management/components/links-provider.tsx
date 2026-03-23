import React, { createContext, useContext, useState } from 'react'
import { Link, LinkGroup } from '../data/schema'

type LinksDialogType = 'add' | 'edit' | 'delete' | 'add-group' | 'edit-group' | 'delete-group'

interface LinksContextType {
  open: LinksDialogType | null
  setOpen: (open: LinksDialogType | null) => void
  currentRow: Link | null
  setCurrentRow: (row: Link | null) => void
  currentGroup: LinkGroup | null
  setCurrentGroup: (group: LinkGroup | null) => void
  groupIndex: number | null
  setGroupIndex: (index: number | null) => void
  linkIndex: number | null
  setLinkIndex: (index: number | null) => void
}

const LinksContext = createContext<LinksContextType | null>(null)

interface Props {
  children: React.ReactNode
}

export default function LinksProvider({ children }: Props) {
  const [open, setOpen] = useState<LinksDialogType | null>(null)
  const [currentRow, setCurrentRow] = useState<Link | null>(null)
  const [currentGroup, setCurrentGroup] = useState<LinkGroup | null>(null)
  const [groupIndex, setGroupIndex] = useState<number | null>(null)
  const [linkIndex, setLinkIndex] = useState<number | null>(null)

  return (
    <LinksContext.Provider
      value={{
        open,
        setOpen,
        currentRow,
        setCurrentRow,
        currentGroup,
        setCurrentGroup,
        groupIndex,
        setGroupIndex,
        linkIndex,
        setLinkIndex,
      }}
    >
      {children}
    </LinksContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useLinks = () => {
  const context = useContext(LinksContext)
  if (!context) {
    throw new Error('useLinks must be used within a LinksProvider')
  }
  return context
}
