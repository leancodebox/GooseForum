import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'
import { useAvatarCrop } from '../src/site/settings/use-avatar-crop'
import { GooseRuntimeProvider, type GooseRuntime } from '@gooseforum/runtime'

const mocks = vi.hoisted(() => ({ change: vi.fn(), destroy: vi.fn(), center: vi.fn(), created: vi.fn() }))
vi.mock('cropperjs', () => ({ default: class {
  container: HTMLElement
  constructor(image: HTMLElement) { this.container = image.parentElement!; mocks.created() }
  getCropperImage() { return { $ready: async () => {}, $center: mocks.center, getBoundingClientRect: () => ({ left: 0, top: 0, width: 300, height: 200 }) } }
  getCropperCanvas() { return { clientWidth: 300, clientHeight: 300, getBoundingClientRect: () => ({ left: 0, top: 0 }) } }
  getCropperSelection() { return { $change: mocks.change, $toCanvas: async () => ({ toDataURL: () => 'preview' }) } }
  destroy() { mocks.destroy() }
} }))
afterEach(() => { cleanup(); vi.restoreAllMocks(); vi.unstubAllGlobals(); vi.clearAllMocks() })

it('initializes after the portal image mounts, creates a default square, and cleans up on close', async () => {
  vi.stubGlobal('URL', { createObjectURL: () => 'blob:test', revokeObjectURL: vi.fn() })
  let crop!: ReturnType<typeof useAvatarCrop>
  function Fixture({ mounted }: { mounted: boolean }) {
    crop = useAvatarCrop({ initialUrl: '', onSuccess: vi.fn(), onError: vi.fn() })
    return <div>{mounted && crop.open ? <img ref={crop.imageRef} alt="crop" /> : null}<output>{String(crop.ready)}</output></div>
  }
  const runtime = { api: { uploads: { avatar: vi.fn() } } } as unknown as GooseRuntime
  const view = (mounted: boolean) => <GooseRuntimeProvider runtime={runtime}><Fixture mounted={mounted} /></GooseRuntimeProvider>
  const { rerender } = render(view(false))
  act(() => crop.selectFile(new File(['image'], 'avatar.png', { type: 'image/png' })))
  await act(async () => { await crop.upload() })
  expect(crop.cropError).toBe('')
  expect(mocks.created).not.toHaveBeenCalled()
  rerender(view(true))
  await waitFor(() => expect(screen.getByRole('status').textContent).toBe('true'))
  expect(mocks.center).toHaveBeenCalledWith('contain')
  expect(mocks.change).toHaveBeenCalledWith(50, 0, 200, 200, 1, true)
  act(() => crop.close())
  expect(mocks.destroy).toHaveBeenCalledOnce()
})
