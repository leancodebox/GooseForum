import { ColumnDef } from '@tanstack/react-table'
import { Category } from '../data/schema'
import { DataTableColumnHeader } from '@/components/data-table/column-header'
import { DataTableRowActions } from './data-table-row-actions'

export const categoriesColumns: ColumnDef<Category>[] = [
  {
    accessorKey: 'id',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='ID' />
    ),
    cell: ({ row }) => <div className='w-[80px]'>{row.getValue('id')}</div>,
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: 'category',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='分类名称' />
    ),
    cell: ({ row }) => {
      const color = row.original.color
      const icon = row.original.icon
      return (
        <div className='flex items-center gap-2'>
          {icon ? (
            <span style={{ color }}>{icon}</span>
          ) : (
            <span
              className='w-2.5 h-2.5 rounded-[3px] shadow-sm'
              style={{ backgroundColor: color }}
            />
          )}
          <span className='max-w-[500px] truncate font-medium'>
            {row.getValue('category')}
          </span>
        </div>
      )
    },
  },
  {
    accessorKey: 'slug',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Slug' />
    ),
    cell: ({ row }) => {
      return (
        <span className='max-w-[500px] truncate text-muted-foreground'>
          {row.getValue('slug') || '-'}
        </span>
      )
    },
  },
  {
    accessorKey: 'desc',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='描述' />
    ),
    cell: ({ row }) => {
      return (
        <span className='max-w-[500px] truncate text-muted-foreground'>
          {row.getValue('desc') || '-'}
        </span>
      )
    },
  },
  {
    id: 'actions',
    cell: ({ row }) => <DataTableRowActions row={row} />,
  },
]
