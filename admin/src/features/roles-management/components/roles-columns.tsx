import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { type ColumnDef, type Row } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { roleSchema, type Role } from '../data/schema'
import { useRoles } from './roles-provider'
import { DataTableColumnHeader } from '@/components/data-table'
import { Badge } from '@/components/ui/badge'

interface DataTableRowActionsProps<TData> {
  row: Row<TData>
}

export function DataTableRowActions<TData>({
  row,
}: DataTableRowActionsProps<TData>) {
  const role = roleSchema.parse(row.original)
  const { setOpen, setCurrentRow } = useRoles()

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          className='flex h-8 w-8 p-0 data-[state=open]:bg-muted'
        >
          <DotsHorizontalIcon className='h-4 w-4' />
          <span className='sr-only'>Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[160px]'>
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(role)
            setOpen('edit')
          }}
        >
          编辑
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(role)
            setOpen('delete')
          }}
          className='text-destructive focus:text-destructive'
        >
          删除
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export const rolesColumns: ColumnDef<Role>[] = [
  {
    accessorKey: 'roleId',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='ID' />
    ),
    cell: ({ row }) => <div className='w-[80px]'>{row.getValue('roleId')}</div>,
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: 'roleName',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='角色名称' />
    ),
    cell: ({ row }) => {
      return (
        <span className='max-w-[500px] truncate font-medium'>
          {row.getValue('roleName')}
        </span>
      )
    },
  },
  {
    accessorKey: 'effective',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='状态' />
    ),
    cell: ({ row }) => {
      const effective = row.getValue('effective') as number
      return (
        <Badge variant={effective === 1 ? 'default' : 'secondary'}>
          {effective === 1 ? '有效' : '无效'}
        </Badge>
      )
    },
    filterFn: (row, id, value) => {
      return value.includes(row.getValue(id))
    },
  },
  {
    accessorKey: 'permissions',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='权限' />
    ),
    cell: ({ row }) => {
      const permissions = row.original.permissions
      return (
        <div className='flex flex-wrap gap-1'>
          {permissions.map((p) => (
            <Badge key={p.id} variant='outline'>
              {p.name}
            </Badge>
          ))}
        </div>
      )
    },
  },
  {
    accessorKey: 'createTime',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='创建时间' />
    ),
    cell: ({ row }) => <div>{row.getValue('createTime')}</div>,
  },
  {
    id: 'actions',
    cell: ({ row }) => <DataTableRowActions row={row} />,
  },
]
