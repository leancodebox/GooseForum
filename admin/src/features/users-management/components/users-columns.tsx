import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import { type User } from '../data/schema'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { DataTableRowActions } from '@/features/users-management/components/data-table-row-actions'

export const usersColumns: ColumnDef<User>[] = [
  {
    accessorKey: 'username',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='用户信息' />
    ),
    cell: ({ row }) => {
      const user = row.original
      return (
        <div className='flex items-center gap-3'>
          <Avatar className='h-10 w-10'>
            <AvatarImage src={user.avatarUrl || undefined} alt={user.username} />
            <AvatarFallback>{user.username.substring(0, 2).toUpperCase()}</AvatarFallback>
          </Avatar>
          <div className='flex flex-col'>
            <span className='font-medium'>{user.username}</span>
            <span className='text-xs text-muted-foreground'>{user.email}</span>
          </div>
        </div>
      )
    },
  },
  {
    accessorKey: 'roleList',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='角色' />
    ),
    cell: ({ row }) => {
      const { roleList } = row.original
      return (
        <div className='flex flex-wrap gap-1'>
          {roleList?.map((role) => (
            <Badge key={role.value} variant='secondary' className='text-[10px]'>
              {role.name}
            </Badge>
          ))}
          {(!roleList || roleList.length === 0) && <span className='text-xs text-muted-foreground'>无</span>}
        </div>
      )
    },
  },
  {
    accessorKey: 'status',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='状态' />
    ),
    cell: ({ row }) => {
      const status = row.getValue('status') as number
      return (
        <Badge variant={status === 1 ? 'destructive' : 'default'} className='text-[10px]'>
          {status === 0 ? '正常' : '已封禁'}
        </Badge>
      )
    },
  },
  {
    accessorKey: 'validate',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='验证' />
    ),
    cell: ({ row }) => {
      const validate = row.getValue('validate') as number
      return (
        <Badge variant={validate === 1 ? 'secondary' : 'outline'} className='text-[10px]'>
          {validate === 1 ? '已验证' : '未验证'}
        </Badge>
      )
    },
  },
  {
    accessorKey: 'createTime',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='注册时间' />
    ),
    cell: ({ row }) => (
      <div className='text-xs text-muted-foreground'>{row.getValue('createTime')}</div>
    ),
  },
  {
    accessorKey: 'lastActiveTime',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='最后登录' />
    ),
    cell: ({ row }) => (
      <div className='text-xs text-muted-foreground'>{row.getValue('lastActiveTime') || '从未登录'}</div>
    ),
  },
  {
    id: 'actions',
    cell: ({ row }) => <DataTableRowActions row={row} />,
  },
]
