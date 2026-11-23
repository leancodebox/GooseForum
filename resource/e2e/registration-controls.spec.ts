import AxeBuilder from '@axe-core/playwright';
import type { Page } from '@playwright/test';
import { expect, test } from './fixtures';

async function mockSecuritySettings(page: Page) {
  await page.route('**/__goose_page/**', route => route.fulfill({ json: {
    component: 'admin.shell', props: {}, url: '/admin/settings/security?lang=en', version: '1.0', meta: { title: 'Security' },
    layout: { site: { name: 'GooseForum', description: '', logo: '', favicon: '', brandType: 'default', brandText: '', brandImage: '' }, viewer: { id: 1, username: 'admin', email: 'admin@example.com', avatarUrl: '', isAuthenticated: true, canAccessAdmin: true, isModerator: false, requiresEmailVerification: false, adminPermissions: [0] }, header: [], sidebar: { activeKey: 'topics', categories: [] }, footer: { links: [], primary: [] }, unread: { notifications: false, messages: false }, theme: { enabled: false, current: 'gf-light', themeColor: '#fbfdff' } },
  } }));
  await page.route('**/api/admin/security-settings', route => route.fulfill({ json: { code: 0, result: { enableSignup: true, enableEmailVerification: true, allowedDomains: ['example.com'], registrationIPLimit: 5, registrationEmailLimit: 3, registrationGlobalLimit: 100 } } }));
}

test('edits hourly registration controls in the flat security settings', async ({ page }, testInfo) => {
  await mockSecuritySettings(page);
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

test('configures a hosted registration captcha', async ({ page }) => {
  await mockSecuritySettings(page);
  let saved: Record<string, unknown> | undefined;
  await page.route('**/api/admin/save-security-settings', route => {
    saved = route.request().postDataJSON();
    return route.fulfill({ json: { code: 0, result: true } });
  });
  await page.goto('/admin/settings/security?lang=en');
  const captcha = page.getByRole('combobox', { name: 'Login and registration verification' });
  await expect(captcha).toHaveText('Numeric image CAPTCHA');
  await captcha.click();
  await expect(page.getByRole('option')).toHaveCount(6);
  await page.getByRole('option', { name: 'Proof of work (ALTCHA)' }).click();
  const difficulty = page.getByRole('combobox', { name: 'Verification difficulty' });
  await expect(difficulty).toHaveText('Medium');
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await expect.poll(() => saved).toMatchObject({ settings: { registrationCaptcha: { provider: 'altcha', difficulty: 'medium' } } });
  await difficulty.click();
  await expect(page.getByRole('option')).toHaveText(['Low', 'Medium', 'High']);
  await page.getByRole('option', { name: 'High' }).click();
  await expect(page.getByLabel('Site Key', { exact: true })).toHaveCount(0);
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await expect.poll(() => saved).toMatchObject({ settings: { registrationCaptcha: { provider: 'altcha', difficulty: 'high' } } });
  await captcha.click();
  await page.getByRole('option', { name: 'reCAPTCHA v3 (score)' }).click();
  await expect(page.getByLabel('Secret Key', { exact: true })).toHaveAccessibleDescription('Enter the secret from your verification provider.');
  await expect(difficulty).toHaveCount(0);
  await page.getByLabel('Site Key', { exact: true }).fill('public-key');
  await page.getByLabel('Secret Key', { exact: true }).fill('test-secret');
  await page.getByLabel('Minimum passing score').fill('0.7');
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await expect.poll(() => saved).toMatchObject({ settings: { registrationCaptcha: { provider: 'recaptcha_v3', siteKey: 'public-key', secret: 'test-secret', scoreThreshold: 0.7 } } });
  await expect(page.getByLabel('Secret Key', { exact: true })).toHaveAccessibleDescription('Already configured. Leave blank to keep the current secret.');
  await page.getByLabel('Site Key', { exact: true }).fill('another-key');
  await expect(page.getByLabel('Secret Key', { exact: true })).toHaveAccessibleDescription('Enter the secret from your verification provider.');
  await captcha.click();
  await page.getByRole('option', { name: 'No CAPTCHA' }).click();
  await expect(page.getByLabel('Secret Key', { exact: true })).toHaveCount(0);
});
