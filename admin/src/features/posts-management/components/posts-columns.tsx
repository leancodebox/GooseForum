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
import type { Category } from '@/api/types'

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
        <DropdownMenuItem onClick={() => { setCurrentRow(post); setOpen('categories'); }}>
          修改分类
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

type CategoryMap = Map<number, Pick<Category, 'id' | 'category' | 'color'>>

const articleTypes: Record<number, { label: string; className: string }> = {
  0: { label: '博文', className: 'bg-blue-50 text-blue-700 ring-blue-100' },
  1: { label: '分享', className: 'bg-emerald-50 text-emerald-700 ring-emerald-100' },
  2: { label: '问答', className: 'bg-amber-50 text-amber-700 ring-amber-100' },
  3: { label: '教程', className: 'bg-violet-50 text-violet-700 ring-violet-100' },
}

function typeInfo(type: number) {
  return articleTypes[type] || { label: '文章', className: 'bg-slate-50 text-slate-700 ring-slate-100' }
}

export function getPostsColumns(categoryMap: CategoryMap): ColumnDef<Post>[] {
  return [
  {
    accessorKey: 'title',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='帖子信息' />
    ),
    cell: ({ row }) => {
      const post = row.original
      const info = typeInfo(post.type)
      const categories = post.categoryId
        .map((id) => categoryMap.get(id))
        .filter(Boolean)
      const createdAt = String(post.createdAt || '')
      const createdDate = createdAt.slice(0, 10)

      return (
        <div className='flex min-w-0 max-w-[640px] flex-col gap-1 pr-3'>
          <div className='flex min-w-0 flex-wrap items-center gap-x-1.5 gap-y-1'>
            <a 
              href={`/p/post/${post.id}`} 
              target="_blank" 
              className='min-w-0 max-w-full truncate text-[15px] font-bold leading-5 text-foreground hover:text-primary hover:underline'
            >
              {post.title}
            </a>
            <span className={`inline-flex h-5 shrink-0 items-center rounded-full px-1.5 text-[11px] font-semibold ring-1 ${info.className}`}>
              {info.label}
            </span>
            {post.processStatus === 1 && (
              <Badge variant='destructive' className='h-5 shrink-0 rounded-full px-1.5 text-[10px] font-semibold'>封禁</Badge>
            )}
            <span className='inline-flex shrink-0 items-center gap-1'>
              {categories.length > 0 ? (
                categories.slice(0, 2).map((category) => (
                <span
                  key={category!.id}
                  className='inline-flex h-5 items-center gap-1 rounded-full bg-muted px-1.5 text-[11px] font-medium text-muted-foreground'
                >
                  <span
                    className='h-1.5 w-1.5 rounded-full'
                    style={{ backgroundColor: category!.color || '#64748b' }}
                  />
                  {category!.category}
                </span>
                ))
              ) : (
                <span className='inline-flex h-5 items-center rounded-full bg-muted px-1.5 text-[11px] text-muted-foreground'>未分类</span>
              )}
            </span>
            <span className='inline-flex shrink-0 items-center gap-2 text-xs text-muted-foreground'>
              <span className='inline-flex items-center gap-1' title='浏览量'>
              <EyeOpenIcon className='h-3.5 w-3.5' />
              <span className='tabular-nums'>{post.viewCount}</span>
              </span>
              <span className='inline-flex items-center gap-1' title='评论数'>
                <ChatBubbleIcon className='h-3.5 w-3.5' />
                <span className='tabular-nums'>{post.replyCount}</span>
              </span>
              <span className='inline-flex items-center gap-1' title='点赞数'>
                <HeartIcon className='h-3.5 w-3.5' />
                <span className='tabular-nums'>{post.likeCount}</span>
              </span>
            </span>
            <span className='shrink-0 text-xs text-muted-foreground/80 xl:hidden'>{createdDate}</span>
          </div>
          <div className='min-w-0 truncate text-xs leading-5 text-muted-foreground'>
            {post.description || '暂无摘要'}
          </div>
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
          <Avatar className='h-8 w-8 ring-1 ring-border'>
            <AvatarImage src={post.userAvatarUrl || ''} alt={post.username} />
            <AvatarFallback>{post.username[0]}</AvatarFallback>
          </Avatar>
          <div className='min-w-0'>
            <a href={`/u/${post.userId}`} target="_blank" className='block max-w-[120px] truncate text-sm font-semibold hover:text-primary hover:underline'>
              {post.username}
            </a>
          </div>
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
        <div className='flex items-center justify-center gap-1.5'>
          <Badge variant={status === 1 ? 'default' : 'secondary'} className='h-6 rounded-md px-2'>
            {status === 1 ? '发布' : '草稿'}
          </Badge>
        </div>
      )
    },
  },
  {
    accessorKey: 'createdAt',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='发布时间' className='hidden w-[92px] xl:flex' />
    ),
    cell: ({ row }) => {
      const createdAt = String(row.getValue('createdAt') || '')
      return <div className='hidden whitespace-nowrap text-xs text-muted-foreground xl:block'>{createdAt.slice(0, 10)}</div>
    },
  },
  {
    id: 'actions',
    cell: ({ row }) => <DataTableRowActions row={row} />,
  },
]
}

export const postsColumns: ColumnDef<Post>[] = getPostsColumns(new Map())
