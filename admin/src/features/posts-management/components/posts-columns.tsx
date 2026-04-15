import { DotsHorizontalIcon, EyeOpenIcon, ChatBubbleIcon, HeartIcon } from '@radix-ui/react-icons'
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
      const types: Record<number, { label: string; color: string }> = {
        0: { label: '博文', color: 'bg-blue-100 text-blue-800 border-blue-200' },
        1: { label: '分享', color: 'bg-green-100 text-green-800 border-green-200' },
        2: { label: '问答', color: 'bg-yellow-100 text-yellow-800 border-yellow-200' }
      }
      const typeInfo = types[post.type] || { label: '文章', color: 'bg-gray-100 text-gray-800 border-gray-200' }

      return (
        <div className='w-[200px] sm:w-[280px] lg:w-[350px] xl:w-[450px] flex flex-col pr-4'>
          <div className='flex items-start gap-2'>
            <Badge variant="outline" className={`mt-0.5 px-1.5 py-0 text-[10px] font-semibold whitespace-nowrap shrink-0 ${typeInfo.color}`}>
              {typeInfo.label}
            </Badge>
            <a 
              href={`/p/post/${post.id}`} 
              target="_blank" 
              className='font-medium text-sm line-clamp-2 hover:underline break-all'
            >
              {post.title}
            </a>
            {post.processStatus === 1 && (
              <Badge variant='destructive' className='mt-0.5 px-1.5 py-0 text-[10px] font-semibold whitespace-nowrap shrink-0'>封禁</Badge>
            )}
          </div>
          <span className='text-xs text-muted-foreground line-clamp-1 mt-1 break-all'>
            {post.description || '暂无摘要'}
          </span>
        </div>
      )
    },
  },
  {
    accessorKey: 'username',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='作者' className='w-[140px] justify-center' />
    ),
    cell: ({ row }) => {
      const post = row.original
      return (
        <div className='flex items-center justify-center gap-2'>
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
    accessorKey: 'articleStatus',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='状态' className='w-[100px] justify-center' />
    ),
    cell: ({ row }) => {
      const status = row.getValue('articleStatus') as number
      return (
        <div className='flex items-center justify-center gap-1.5 flex-wrap'>
          <Badge variant={status === 1 ? 'default' : 'secondary'}>
            {status === 1 ? '发布' : '草稿'}
          </Badge>
        </div>
      )
    },
  },
  {
    id: 'stats',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='统计' className='w-[160px] justify-center' />
    ),
    cell: ({ row }) => {
      const post = row.original
      return (
        <div className='flex items-center justify-center gap-4 text-muted-foreground whitespace-nowrap w-full'>
          <div className='flex items-center gap-1.5' title='浏览量'>
            <EyeOpenIcon className='h-3.5 w-3.5' />
            <span className='text-xs tabular-nums min-w-[20px]'>{post.viewCount}</span>
          </div>
          <div className='flex items-center gap-1.5' title='评论数'>
            <ChatBubbleIcon className='h-3 w-3' />
            <span className='text-xs tabular-nums min-w-[20px]'>{post.replyCount}</span>
          </div>
          <div className='flex items-center gap-1.5' title='点赞数'>
            <HeartIcon className='h-3.5 w-3.5' />
            <span className='text-xs tabular-nums min-w-[20px]'>{post.likeCount}</span>
          </div>
        </div>
      )
    },
  },
  {
    accessorKey: 'createdAt',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='发布时间' className='w-[140px]' />
    ),
    cell: ({ row }) => <div className='text-xs whitespace-nowrap'>{row.getValue('createdAt')}</div>,
  },
  {
    id: 'actions',
    cell: ({ row }) => <DataTableRowActions row={row} />,
  },
]
