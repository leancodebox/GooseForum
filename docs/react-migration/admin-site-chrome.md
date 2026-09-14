# Admin site chrome migration record

- React route: `/admin/settings/site-chrome`
- Vue reference: `SiteChromeManagementPage.vue`
- Preserved configuration: brand type/text/image, header links, main menu, resources, custom sidebar groups/items, category preview, footer links and primary text.
- Item lists retain drag ordering through the existing dnd-kit dependency; system header defaults are restored when absent.
- Brand image upload and the complete chrome configuration use the framework-neutral admin client.
- No live chrome settings were changed during verification.
- Follow-up fidelity work restored the C-side fixed primary/resource entries, 224px sidebar density, category/footer ordering, header system-item visibility restriction, custom-item deletion, and the original header/content preview proportions.
