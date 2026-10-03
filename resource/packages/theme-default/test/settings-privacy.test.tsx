import { afterEach, expect, it, vi } from 'vitest'
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { GooseI18nProvider } from '@gooseforum/runtime/i18n'
import { PrivacySettings } from '../src/site/settings/settings-privacy'

afterEach(cleanup)

it('renders saved values and disables all privacy controls while saving', async () => {
  const onChange = vi.fn()
  const settings = { showActivity: true, showTopics: false, showFollowing: true }
  const view = render(<GooseI18nProvider locale="en">
    <PrivacySettings settings={settings} saving={false} onChange={onChange} />
  </GooseI18nProvider>)
  const user = userEvent.setup()
  expect(screen.getByRole('checkbox', { name: 'Show my topics' }).getAttribute('aria-checked')).toBe('false')
  await user.click(screen.getByRole('checkbox', { name: 'Show my topics' }))
  expect(onChange).toHaveBeenCalledWith('showTopics', true)
  view.rerender(<GooseI18nProvider locale="en">
    <PrivacySettings settings={settings} saving={true} onChange={onChange} />
  </GooseI18nProvider>)
  for (const checkbox of screen.getAllByRole('checkbox')) expect(checkbox.hasAttribute('disabled')).toBe(true)
  await user.click(screen.getByRole('checkbox', { name: 'Show follow relationships' }))
  expect(onChange).toHaveBeenCalledOnce()
})
