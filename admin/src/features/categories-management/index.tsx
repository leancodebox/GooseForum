import { useQuery } from '@tanstack/react-query'
import { PlusIcon } from '@radix-ui/react-icons'
import { ContentLayout } from '@/components/layout/content-layout'
import { Button } from '@/components/ui/button'
import { getCategoryList } from '@/api'
import { CategoriesProvider, useCategories } from './components/categories-provider'
import { categoriesColumns } from './components/categories-columns'
import { CategoriesActionDialog } from './components/categories-action-dialog'
import { CategoriesDeleteDialog } from './components/categories-delete-dialog'
import { CategoriesTable } from './components/categories-table'

function CategoriesManagementContent() {
  const { setOpen } = useCategories()

  const { data, isLoading, refetch } = useQuery<any>({
    queryKey: ['categories'],
    queryFn: () => getCategoryList(),
  })

  return (
    <>
      <ContentLayout
        title='分类管理'
        description='管理论坛的所有文章分类。'
        headerActions={
          <Button onClick={() => setOpen('create')}>
            <PlusIcon className='mr-2 h-4 w-4' /> 新增分类
          </Button>
        }
      >
        <div className='-mx-4 flex-1 overflow-auto px-4 py-1 lg:flex-row lg:space-x-12 lg:space-y-0'>
          <CategoriesTable
            columns={categoriesColumns}
            data={data?.result || []}
            loading={isLoading}
          />
        </div>
      </ContentLayout>

      <CategoriesActionDialog onSuccess={() => refetch()} />
      <CategoriesDeleteDialog onSuccess={() => refetch()} />
    </>
  )
}

export default function CategoriesManagement() {
  return (
    <CategoriesProvider>
      <CategoriesManagementContent />
    </CategoriesProvider>
  )
}
