import { useState } from 'react'
import { toast } from 'sonner'
import axios from 'axios'
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
import { useCategories } from './categories-provider'

interface CategoriesDeleteDialogProps {
  onSuccess: () => void
}

export function CategoriesDeleteDialog({ onSuccess }: CategoriesDeleteDialogProps) {
  const { open, setOpen, currentRow } = useCategories()
  const [loading, setLoading] = useState(false)

  const onDelete = async () => {
    if (!currentRow) return
    setLoading(true)
    try {
      const response = await axios.post('/api/admin/category-delete', {
        id: currentRow.id,
      })

      if (response.data.code === 0) {
        toast.success('删除成功')
        setOpen(null)
        onSuccess()
      } else {
        toast.error(response.data.message || '删除失败')
      }
    } catch (error) {
      console.error('Failed to delete category:', error)
      toast.error('删除失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <AlertDialog open={open === 'delete'} onOpenChange={(v) => !v && setOpen(null)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确认删除</AlertDialogTitle>
          <AlertDialogDescription>
            您确定要删除分类 “{currentRow?.category}” 吗？此操作不可撤销，且至少需保留一个分类。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={loading}>取消</AlertDialogCancel>
          <AlertDialogAction
            disabled={loading}
            className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
            onClick={(e) => {
              e.preventDefault()
              onDelete()
            }}
          >
            {loading && <span className='mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent' />}
            确认删除
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
