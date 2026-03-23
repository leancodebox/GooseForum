import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ContentLayout } from '@/components/layout/content-layout'
import { Button } from '@/components/ui/button'
import { PlusIcon, ReloadIcon } from '@radix-ui/react-icons'
import { getArticlesList } from '@/api'
import { PostsProvider } from './components/posts-provider'
import { postsColumns } from './components/posts-columns'
import { PostsTable } from './components/posts-table'
import { PostsActionDialog } from './components/posts-action-dialog'

function PostsManagementContent() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [search, setSearch] = useState('')

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['articles', page, pageSize, search],
    queryFn: () => getArticlesList({ page, pageSize, search }),
  })

  return (
    <>
      <ContentLayout
        title='帖子管理'
        description='管理系统中的所有帖子内容。'
        headerActions={
          <div className='flex items-center space-x-2'>
            <Button variant='outline' size='sm' onClick={() => refetch()} disabled={isLoading}>
              {isLoading ? <ReloadIcon className='mr-2 h-4 w-4 animate-spin' /> : <ReloadIcon className='mr-2 h-4 w-4' />}
              刷新
            </Button>
            <Button size='sm' onClick={() => window.open('/publish', '_blank')}>
              <PlusIcon className='mr-2 h-4 w-4' />
              发布帖子
            </Button>
          </div>
        }
      >
        <div className='-mx-4 flex-1 overflow-auto px-4 py-1 lg:flex-row lg:space-x-12 lg:space-y-0'>
          <PostsTable
            columns={postsColumns}
            data={data?.result?.list || []}
            total={data?.result?.total || 0}
            page={page}
            pageSize={pageSize}
            onPageChange={setPage}
            onPageSizeChange={setPageSize}
            onSearchChange={(val) => {
              setSearch(val)
              setPage(1)
            }}
          />
        </div>
      </ContentLayout>
      <PostsActionDialog />
    </>
  )
}

export default function PostsManagement() {
  return (
    <PostsProvider>
      <PostsManagementContent />
    </PostsProvider>
  )
}
