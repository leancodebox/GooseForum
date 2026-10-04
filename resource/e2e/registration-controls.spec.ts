import AxeBuilder from '@axe-core/playwright';
import { expect, test } from '@playwright/test';

test('edits hourly registration controls in the flat security settings', async ({ page }, testInfo) => {
  await page.route('**/__goose_page/**', route => route.fulfill({ json: {
    component: 'admin.shell', props: {}, url: '/admin/settings/security?lang=en', version: '1.0', meta: { title: 'Security' },
    layout: { site: { name: 'GooseForum', description: '', logo: '', favicon: '', brandType: 'default', brandText: '', brandImage: '' }, viewer: { id: 1, username: 'admin', email: 'admin@example.com', avatarUrl: '', isAuthenticated: true, canAccessAdmin: true, isModerator: false, requiresEmailVerification: false, adminPermissions: [0] }, header: [], sidebar: { activeKey: 'topics', categories: [] }, footer: { links: [], primary: [] }, unread: { notifications: false, messages: false }, theme: { enabled: false, current: 'gf-light', themeColor: '#fbfdff' } },
  } }));
  await page.route('**/api/admin/security-settings', route => route.fulfill({ json: { code: 0, result: { enableSignup: true, enableEmailVerification: true, allowedDomains: ['example.com'], registrationIPLimit: 5, registrationEmailLimit: 3, registrationGlobalLimit: 100 } } }));
  let saved: Record<string, unknown> | undefined;
  await page.route('**/api/admin/save-security-settings', route => {
    saved = route.request().postDataJSON();
    return route.fulfill({ json: { code: 0, result: true } });
  });
  await page.goto('/admin/settings/security?lang=en');
  await expect(page.getByRole('heading', { name: 'Registration limits' })).toBeVisible();
  await page.getByLabel('Per IP address').fill('0');
  await page.getByLabel('All registrations').fill('200');
  expect(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1)).toBe(false);
  await testInfo.attach('registration-security-settings', { body: await page.screenshot({ fullPage: true, path: `/tmp/gooseforum-registration-${testInfo.project.name}.png` }), contentType: 'image/png' });
  const accessibility = await new AxeBuilder({ page }).include('main').withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa']).analyze();
  expect(accessibility.violations).toEqual([]);
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await expect.poll(() => saved).toMatchObject({ settings: { registrationIPLimit: 0, registrationEmailLimit: 3, registrationGlobalLimit: 200 } });
});
