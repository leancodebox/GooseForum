import { createContext, useContext, useState } from 'react'
import { Role } from '../data/schema'
import useDialogState from '@/hooks/use-dialog-state'

type RolesDialogType = 'edit' | 'add' | 'delete'

interface RolesContextType {
  open: RolesDialogType | null
  setOpen: (str: RolesDialogType | null) => void
  currentRow: Role | null
  setCurrentRow: (row: Role | null) => void
}

const RolesContext = createContext<RolesContextType | null>(null)

interface RolesProviderProps {
  children: React.ReactNode
}

export function RolesProvider({ children }: RolesProviderProps) {
  const [open, setOpen] = useDialogState<RolesDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Role | null>(null)

  return (
    <RolesContext.Provider value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </RolesContext.Provider>
  )
}

export const useRoles = () => {
  const context = useContext(RolesContext)
  if (!context) {
    throw new Error('useRoles must be used within a RolesProvider')
  }
  return context
}
