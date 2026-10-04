import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import type { GooseAdminApi } from '@gooseforum/client';
import { SecuritySettingsPage } from './security-settings-page';

afterEach(cleanup);

it('preserves an explicitly disabled registration quota when saving', async () => {
  const saveSecurity = vi.fn().mockResolvedValue(undefined);
  const api = { settings: { security: vi.fn().mockResolvedValue({ enableSignup: true, enableEmailVerification: false, allowedDomains: ['example.com'], registrationIPLimit: 0, registrationEmailLimit: 3, registrationGlobalLimit: 100 }), saveSecurity } } as unknown as GooseAdminApi;
  render(<SecuritySettingsPage api={api} text={(key) => key} />);
  await waitFor(() => expect((screen.getByLabelText('registrationIPLimit') as HTMLInputElement).value).toBe('0'));
  fireEvent.change(screen.getByLabelText('registrationGlobalLimit'), { target: { value: '200' } });
  fireEvent.click(screen.getByRole('button', { name: 'save' }));
  await waitFor(() => expect(saveSecurity).toHaveBeenCalledWith(expect.objectContaining({ registrationIPLimit: 0, registrationEmailLimit: 3, registrationGlobalLimit: 200 })));
});
