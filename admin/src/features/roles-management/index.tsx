import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ContentLayout } from '@/components/layout/content-layout'
import { Button } from '@/components/ui/button'
import { PlusIcon } from '@radix-ui/react-icons'
import { getRoleList } from '@/api'
import { RolesProvider, useRoles } from './components/roles-provider'
import { rolesColumns } from './components/roles-columns'
import { RolesTable } from './components/roles-table'
import { RolesActionDialog } from './components/roles-action-dialog'
import { RolesDeleteDialog } from './components/roles-delete-dialog'

function RolesManagementContent() {
  const { setOpen, setCurrentRow } = useRoles()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [roleName, setRoleName] = useState('')

  const { data } = useQuery({
    queryKey: ['roles', page, pageSize, roleName],
    queryFn: () => getRoleList(),
  })

  return (
    <>
      <ContentLayout
        title='角色管理'
        description='在这里管理系统的角色及其权限。'
        headerActions={
          <Button
            onClick={() => {
              setCurrentRow(null)
              setOpen('add')
            }}
          >
            <PlusIcon className='mr-2 h-4 w-4' />
            添加角色
          </Button>
        }
      >
        <div className='-mx-4 flex-1 overflow-auto px-4 py-1 lg:flex-row lg:space-x-12 lg:space-y-0'>
          <RolesTable
            columns={rolesColumns}
            data={data?.result?.list || []}
            total={data?.result?.total || 0}
            page={page}
            pageSize={pageSize}
            onPageChange={setPage}
            onPageSizeChange={setPageSize}
            onRoleNameChange={(val) => {
              setRoleName(val)
              setPage(1)
            }}
          />
        </div>
      </ContentLayout>
      <RolesActionDialog />
      <RolesDeleteDialog />
    </>
  )
}

export default function RolesManagement() {
  return (
    <RolesProvider>
      <RolesManagementContent />
    </RolesProvider>
  )
}
