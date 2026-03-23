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
import { postSchema, type Post } from '../data/schema'
import { usePosts } from './posts-provider'
import { DataTableColumnHeader } from '@/components/data-table'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'

interface DataTableRowActionsProps<TData> {
  row: Row<TData>
}

export function DataTableRowActions<TData>({
  row,
}: DataTableRowActionsProps<TData>) {
  const post = postSchema.parse(row.original)
  const { setOpen, setCurrentRow } = usePosts()

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
        <DropdownMenuItem onClick={() => window.open(`/p/post/${post.id}`, '_blank')}>
          查看
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => { setCurrentRow(post); setOpen('top'); }}>
          置顶
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => { setCurrentRow(post); setOpen('recommend'); }}>
          推荐
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        {post.processStatus === 1 ? (
          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(post)
              setOpen('approve')
            }}
            className='text-green-600 focus:text-green-600'
          >
            恢复
          </DropdownMenuItem>
        ) : (
          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(post)
              setOpen('reject')
            }}
            className='text-destructive focus:text-destructive'
          >
            封禁
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export const postsColumns: ColumnDef<Post>[] = [
  {
    accessorKey: 'title',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='帖子信息' />
    ),
    cell: ({ row }) => {
      const post = row.original
      return (
        <div className='flex flex-col min-w-0 max-w-[400px]'>
          <a 
            href={`/p/post/${post.id}`} 
            target="_blank" 
            className='font-medium text-sm line-clamp-2 hover:underline'
          >
            {post.title}
          </a>
          <span className='text-xs text-muted-foreground line-clamp-1 mt-1'>
            {post.description || '暂无摘要'}
          </span>
        </div>
      )
    },
  },
  {
    accessorKey: 'username',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='作者' />
    ),
    cell: ({ row }) => {
      const post = row.original
      return (
        <div className='flex items-center gap-2'>
          <Avatar className='h-7 w-7'>
            <AvatarImage src={post.userAvatarUrl || ''} alt={post.username} />
            <AvatarFallback>{post.username[0]}</AvatarFallback>
          </Avatar>
          <a href={`/u/${post.userId}`} target="_blank" className='text-sm hover:underline'>
            {post.username}
          </a>
        </div>
      )
    },
  },
  {
    accessorKey: 'type',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='分类' />
    ),
    cell: ({ row }) => {
      const type = row.getValue('type') as number
      const types: Record<number, string> = {
        0: '博文',
        1: '教程',
        2: '问答',
        3: '分享'
      }
      return <Badge variant='outline'>{types[type] || '文章'}</Badge>
    },
  },
  {
    accessorKey: 'articleStatus',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='状态' />
    ),
    cell: ({ row }) => {
      const status = row.getValue('articleStatus') as number
      const processStatus = row.original.processStatus
      return (
        <div className='flex flex-col gap-1'>
          <Badge variant={status === 1 ? 'default' : 'secondary'}>
            {status === 1 ? '发布' : '草稿'}
          </Badge>
          {processStatus === 1 && (
            <Badge variant='destructive'>封禁</Badge>
          )}
        </div>
      )
    },
  },
  {
    id: 'stats',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='统计' />
    ),
    cell: ({ row }) => {
      const post = row.original
      return (
        <div className='text-xs space-y-1 text-muted-foreground'>
          <div>浏览: {post.viewCount}</div>
          <div>评论: {post.replyCount}</div>
          <div>点赞: {post.likeCount}</div>
        </div>
      )
    },
  },
  {
    accessorKey: 'createdAt',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='发布时间' />
    ),
    cell: ({ row }) => <div className='text-xs whitespace-nowrap'>{row.getValue('createdAt')}</div>,
  },
  {
    id: 'actions',
    cell: ({ row }) => <DataTableRowActions row={row} />,
  },
]
