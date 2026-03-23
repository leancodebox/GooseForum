import React, { createContext, useContext, useState } from 'react'
import { Category } from '../data/schema'

type CategoriesDialogType = 'create' | 'edit' | 'delete'

interface CategoriesContextType {
  open: CategoriesDialogType | null
  setOpen: (open: CategoriesDialogType | null) => void
  currentRow: Category | null
  setCurrentRow: (row: Category | null) => void
}

const CategoriesContext = createContext<CategoriesContextType | undefined>(undefined)

export const CategoriesProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [open, setOpen] = useState<CategoriesDialogType | null>(null)
  const [currentRow, setCurrentRow] = useState<Category | null>(null)

  return (
    <CategoriesContext.Provider value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </CategoriesContext.Provider>
  )
}

export const useCategories = () => {
  const context = useContext(CategoriesContext)
  if (!context) {
    throw new Error('useCategories must be used within a CategoriesProvider')
  }
  return context
}
