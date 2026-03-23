import React, { createContext, useContext, useState } from 'react'
import { SponsorItem, UserSponsor, Sponsors } from '../data/schema'

type SponsorshipDialogType = 'add-sponsor' | 'edit-sponsor' | 'delete-sponsor' | 'add-user' | 'edit-user' | 'delete-user'

interface SponsorshipContextType {
  open: SponsorshipDialogType | null
  setOpen: (open: SponsorshipDialogType | null) => void
  currentRow: SponsorItem | UserSponsor | null
  setCurrentRow: (row: SponsorItem | UserSponsor | null) => void
  currentLevel: keyof Sponsors | null
  setCurrentLevel: (level: keyof Sponsors | null) => void
  currentIndex: number | null
  setCurrentIndex: (index: number | null) => void
}

const SponsorshipContext = createContext<SponsorshipContextType | null>(null)

interface Props {
  children: React.ReactNode
}

export default function SponsorshipProvider({ children }: Props) {
  const [open, setOpen] = useState<SponsorshipDialogType | null>(null)
  const [currentRow, setCurrentRow] = useState<SponsorItem | UserSponsor | null>(null)
  const [currentLevel, setCurrentLevel] = useState<keyof Sponsors | null>(null)
  const [currentIndex, setCurrentIndex] = useState<number | null>(null)

  return (
    <SponsorshipContext.Provider
      value={{
        open,
        setOpen,
        currentRow,
        setCurrentRow,
        currentLevel,
        setCurrentLevel,
        currentIndex,
        setCurrentIndex,
      }}
    >
      {children}
    </SponsorshipContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useSponsorship = () => {
  const context = useContext(SponsorshipContext)
  if (!context) {
    throw new Error('useSponsorship must be used within a SponsorshipProvider')
  }
  return context
}
