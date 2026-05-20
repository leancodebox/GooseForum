import { type Row } from '@tanstack/react-table'
import { Edit3 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { userSchema } from '../data/schema'
import { useUsers } from './users-provider'

interface DataTableRowActionsProps<TData> {
  row: Row<TData>
}

export function DataTableRowActions<TData>({
  row,
}: DataTableRowActionsProps<TData>) {
  const user = userSchema.parse(row.original)
  const { setOpen, setCurrentRow } = useUsers()

  return (
    <Button
      variant='ghost'
      size='icon'
      className='h-8 w-8'
      title='编辑'
      onClick={() => {
        setCurrentRow(user)
        setOpen('edit')
      }}
    >
      <Edit3 className='h-4 w-4' />
      <span className='sr-only'>编辑</span>
    </Button>
  )
}
