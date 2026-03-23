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
import { usePosts } from './posts-provider'

export function PostsActionDialog() {
  const { open, setOpen, currentRow, setCurrentRow } = usePosts()

  const handleAction = async () => {
    if (!currentRow) return

    try {
      let response;
      if (open === 'approve' || open === 'reject') {
        response = await axios.post('/api/admin/article-edit', {
          id: currentRow.id,
          processStatus: open === 'reject' ? 1 : 0,
        })
      } else if (open === 'top' || open === 'recommend') {
        // 后端暂时没看到对应的 API，Vue 代码里也是 todo
        toast.info('该功能暂未实现')
        setOpen(null)
        return
      }

      if (response && response.data.code === 0) {
        toast.success('操作成功')
        setOpen(null)
        setCurrentRow(null)
        window.location.reload()
      } else {
        toast.error(response?.data.message || '操作失败')
      }
    } catch (error) {
      toast.error('请求失败')
    }
  }

  const getTitle = () => {
    switch (open) {
      case 'approve': return '确认恢复文章？'
      case 'reject': return '确认封禁文章？'
      case 'top': return '确认置顶文章？'
      case 'recommend': return '确认推荐文章？'
      default: return ''
    }
  }

  const getDescription = () => {
    switch (open) {
      case 'approve': return `确认要恢复文章 "${currentRow?.title}" 吗？`
      case 'reject': return `确认要封禁文章 "${currentRow?.title}" 吗？封禁后用户将无法查看。`
      case 'top': return `确认要置顶文章 "${currentRow?.title}" 吗？`
      case 'recommend': return `确认要推荐文章 "${currentRow?.title}" 吗？`
      default: return ''
    }
  }

  return (
    <AlertDialog
      open={open !== null && open !== 'view'}
      onOpenChange={(val) => !val && setOpen(null)}
    >
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{getTitle()}</AlertDialogTitle>
          <AlertDialogDescription>
            {getDescription()}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={() => setCurrentRow(null)}>
            取消
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={handleAction}
            className={open === 'reject' ? 'bg-destructive text-destructive-foreground hover:bg-destructive/90' : ''}
          >
            确认
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
