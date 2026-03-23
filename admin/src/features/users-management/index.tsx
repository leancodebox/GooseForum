'use client'

import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ContentLayout } from '@/components/layout/content-layout'
import { getUserList } from '@/api'
import { usersColumns } from './components/users-columns'
import { UsersTable } from './components/users-table'
import { UsersProvider, useUsers } from './components/users-provider'
import { UsersActionDialog } from './components/users-action-dialog'

function UsersManagementContent() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [username, setUsername] = useState('')

  const { open, setOpen, currentRow, setCurrentRow } = useUsers()

  const { data } = useQuery({
    queryKey: ['users', page, pageSize, username],
    queryFn: () => getUserList({ page, pageSize, username: username || undefined }),
  })

  return (
    <>
      <ContentLayout
        title='用户管理'
        description='在此管理系统用户，您可以修改用户状态、验证状态以及分配角色。'
      >
        <div className='flex flex-1 flex-col space-y-4'>
          <UsersTable
            columns={usersColumns}
            data={data?.result?.list || []}
            total={data?.result?.total || 0}
            page={page}
            pageSize={pageSize}
            onPageChange={setPage}
            onPageSizeChange={setPageSize}
            onUsernameChange={(val) => {
              setUsername(val)
              setPage(1)
            }}
          />
        </div>
      </ContentLayout>

      <UsersActionDialog
        key={currentRow?.userId}
        open={open === 'edit'}
        onOpenChange={(isOpen) => {
          if (!isOpen) {
            setOpen(null)
            setCurrentRow(null)
          }
        }}
        currentRow={currentRow}
        onSuccess={() => queryClient.invalidateQueries({ queryKey: ['users'] })}
      />
    </>
  )
}

export default function UsersManagement() {
  return (
    <UsersProvider>
      <UsersManagementContent />
    </UsersProvider>
  )
}
