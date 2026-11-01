import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'
import { useAvatarCrop } from '../src/site/settings/use-avatar-crop'
import { GooseRuntimeProvider, type GooseRuntime } from '@gooseforum/runtime'

const mocks = vi.hoisted(() => ({ change: vi.fn(), destroy: vi.fn(), center: vi.fn(), created: vi.fn(), toCanvas: vi.fn(), selection: null as EventTarget | null }))
vi.mock('cropperjs', () => ({ default: class {
  container: HTMLElement
  selection = Object.assign(new EventTarget(), { $change: mocks.change, $toCanvas: mocks.toCanvas })
  constructor(image: HTMLElement) { this.container = image.parentElement!; mocks.selection = this.selection; mocks.created() }
  getCropperImage() { return { $ready: async () => {}, $center: mocks.center, getBoundingClientRect: () => ({ left: 0, top: 0, width: 300, height: 200 }) } }
  getCropperCanvas() { return { clientWidth: 300, clientHeight: 300, getBoundingClientRect: () => ({ left: 0, top: 0 }) } }
  getCropperSelection() { return this.selection }
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
  mocks.toCanvas.mockResolvedValue({ toDataURL: () => 'preview' })
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

it('initializes a centered 4:3 selection for profile backgrounds', async () => {
  vi.stubGlobal('URL', { createObjectURL: () => 'blob:test', revokeObjectURL: vi.fn() })
  let crop!: ReturnType<typeof useAvatarCrop>
  function Fixture() {
    crop = useAvatarCrop({ initialUrl: '', onSuccess: vi.fn(), onError: vi.fn(), coverUpload: vi.fn() })
    return <div>{crop.open ? <img ref={crop.imageRef} alt="crop" /> : null}<output>{String(crop.ready)}</output></div>
  }
  mocks.toCanvas.mockResolvedValue({ toDataURL: () => 'preview' })
  render(<GooseRuntimeProvider runtime={{ api: {} } as unknown as GooseRuntime}><Fixture /></GooseRuntimeProvider>)
  act(() => crop.selectFile(new File(['image'], 'cover.png', { type: 'image/png' })))
  await waitFor(() => expect(screen.getByRole('status').textContent).toBe('true'))
  expect(mocks.change).toHaveBeenCalledWith(expect.closeTo(16.6667, 3), expect.closeTo(0, 3), expect.closeTo(266.6667, 3), expect.closeTo(200, 3), 4 / 3, true)
  const permitted = new CustomEvent('change', { cancelable: true, detail: { width: 180, height: 100 } })
  const rejected = new CustomEvent('change', { cancelable: true, detail: { width: 220, height: 100 } })
  act(() => { mocks.selection!.dispatchEvent(permitted); mocks.selection!.dispatchEvent(rejected) })
  expect(permitted.defaultPrevented).toBe(false)
  expect(rejected.defaultPrevented).toBe(true)
  await waitFor(() => expect(mocks.toCanvas).toHaveBeenCalledWith(expect.objectContaining({ width: 160 })))
  expect(mocks.toCanvas.mock.calls.at(-1)?.[0]).not.toHaveProperty('height')
})
