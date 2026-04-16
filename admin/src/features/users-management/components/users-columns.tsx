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
          <a href={`/u/${user.userId}`} target='_blank' rel='noreferrer' className='shrink-0'>
            <Avatar className='h-10 w-10 transition-opacity hover:opacity-85'>
              <AvatarImage src={user.avatarUrl || undefined} alt={user.username} />
              <AvatarFallback>{user.username.substring(0, 2).toUpperCase()}</AvatarFallback>
            </Avatar>
          </a>
          <div className='flex min-w-0 flex-col'>
            <a
              href={`/u/${user.userId}`}
              target='_blank'
              rel='noreferrer'
              className='font-medium hover:underline'
            >
              {user.username}
            </a>
            <div className='flex items-center gap-2 text-xs text-muted-foreground'>
              <span className='truncate'>{user.email}</span>
              <Badge variant={user.validate === 1 ? 'secondary' : 'outline'} className='text-[10px]'>
                {user.validate === 1 ? '邮箱已验证' : '邮箱未验证'}
              </Badge>
            </div>
          </div>
        </div>
      )
    },
  },
  {
    accessorKey: 'roleList',
    enableSorting: false,
    header: () => <div className='w-full text-center'>角色</div>,
    cell: ({ row }) => {
      const { roleList } = row.original
      return (
        <div className='flex flex-wrap justify-center gap-1'>
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
    enableSorting: false,
    header: () => <div className='w-full text-center'>状态</div>,
    cell: ({ row }) => {
      const status = row.getValue('status') as number
      return (
        <div className='flex justify-center'>
          <Badge variant={status === 1 ? 'destructive' : 'default'} className='text-[10px]'>
            {status === 0 ? '正常' : '已封禁'}
          </Badge>
        </div>
      )
    },
  },
  {
    accessorKey: 'createTime',
    enableSorting: false,
    header: () => <div className='w-full text-center'>注册时间</div>,
    cell: ({ row }) => (
      <div className='w-full text-center text-xs text-muted-foreground'>{row.getValue('createTime')}</div>
    ),
  },
  {
    accessorKey: 'lastActiveTime',
    enableSorting: false,
    header: () => <div className='w-full text-center'>最后登录</div>,
    cell: ({ row }) => (
      <div className='w-full text-center text-xs text-muted-foreground'>{row.getValue('lastActiveTime') || '从未登录'}</div>
    ),
  },
  {
    id: 'actions',
    cell: ({ row }) => <DataTableRowActions row={row} />,
  },
]
