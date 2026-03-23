import axios from 'axios'
import { toast } from 'sonner'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { useRoles } from './roles-provider'

export function RolesDeleteDialog() {
  const { open, setOpen, currentRow, setCurrentRow } = useRoles()

  const onDelete = async () => {
    if (!currentRow) return

    try {
      const response = await axios.post('/api/admin/role-del', {
        id: currentRow.roleId,
      })

      if (response.data.code === 0) {
        toast.success('删除成功')
        setOpen(null)
        setCurrentRow(null)
        window.location.reload()
      } else {
        toast.error(response.data.message || '删除失败')
      }
    } catch (error) {
      toast.error('请求失败')
    }
  }

  return (
    <AlertDialog
      open={open === 'delete'}
      onOpenChange={(val) => !val && setOpen(null)}
    >
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确认删除角色？</AlertDialogTitle>
          <AlertDialogDescription>
            此操作将永久删除角色{' '}
            <span className='font-bold'>{currentRow?.roleName}</span>
            ，且不可撤销。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={() => setCurrentRow(null)}>
            取消
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={onDelete}
            className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
          >
            删除
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
