import axiosInstance from './axiosInstance.ts';
import type {
    AdminArticlesItem,
    ApplySheet,
    Category,
    FriendLinksGroup,
    Label,
    Result,
    User,
    UserRole,
    PageData,
    SponsorsConfig
} from './adminInterfaces.ts';


// Mock 获取用户信息
export const getUserInfo = async (): Promise<Result<any>> => {
    return axiosInstance.get('api/get-user-info')
}


export function getUserList(page: number, size: number): Promise<Result<PageData<User>>> {
    return axiosInstance.post('api/admin/user-list', {
        page: page,
        pageSize: size,
    })
}

export function editUser(userId: any, status: any, validate: any, roleId: any): Promise<Result<any>> {
    return axiosInstance.post('api/admin/user-edit', {
        userId: userId,
        status: status,
        validate: validate,
        roleId: roleId,
    })
}

export function getAllRoleItem(): Promise<Result<Label[]>> {
    return axiosInstance.post('api/admin/get-all-role-item')
}

export function getPermissionList(): Promise<Result<Label[]>> {
    return axiosInstance.post('api/admin/get-permission-list')
}

export function getRoleList(): Promise<Result<PageData<UserRole>>> {
    return axiosInstance.post('api/admin/role-list')
}

export function getRoleSave(id: any, roleName: any, permission: any): Promise<Result<any>> {
    return axiosInstance.post('api/admin/role-save', {
        id: id,
        roleName: roleName,
        permissions: permission
    })
}

export function getRoleDel(id: any): Promise<Result<any>> {
    return axiosInstance.post('api/admin/role-delete', {
        id: id,
    })
}

export function getAdminArticlesList(page = 1, pageSize = 10): Promise<Result<PageData<AdminArticlesItem>>> {
    return axiosInstance.post('api/admin/articles-list', {
        page: page,
        pageSize: pageSize,
    })
}

export const getCategoryList = (): Promise<Result<Category[]>> => {
    return axiosInstance.post('api/admin/category-list')
}

export const saveCategory = (id: number, category: string,desc:string, sort: number, status: number): Promise<Result<any>> => {
    return axiosInstance.post('api/admin/category-save', {
        id: id,
        category: category,
        desc:desc,
        sort: sort,
        status: status,
    })
}

export const deleteCategory = (id: number): Promise<Result<any>> => {
    return axiosInstance.post('api/admin/category-delete', {id})
}

export const editArticle = (id: any, processStatus: any): Promise<Result<any>> => {
    return axiosInstance.post('api/admin/article-edit', {
        id,
        processStatus
    })
}

export function applySheetList(page = 1, pageSize = 10): Promise<Result<PageData<ApplySheet>>> {
    return axiosInstance.post('api/admin/apply-sheet-list', {
        page: page,
        pageSize: pageSize,
    })
}


export function getFriendLinks(): Promise<Result<FriendLinksGroup[]>> {
    return axiosInstance.get('api/admin/friend-links')
}

export function saveFriendLinks(params: any): Promise<Result<FriendLinksGroup[]>> {
    return axiosInstance.post('api/admin/save-friend-links', {
        linksInfo: params
    })
}

export function getSiteStatistics(): Promise<Result<any>> {
    return axiosInstance.get('api/forum/get-site-statistics')
}

export const getArticleEnum = async (): Promise<Result<Record<string,Label[]>>> => {
    return axiosInstance.get('api/forum/get-articles-enum');
}

// 网页设置相关接口
export interface WebSettingsConfig {
    externalLinks: string
}

export const getWebSettings = (): Promise<Result<WebSettingsConfig>> => {
    return axiosInstance.get('api/admin/web-settings');
}

export const saveWebSettings = (settings: WebSettingsConfig): Promise<Result<any>> => {
    return axiosInstance.post('api/admin/save-web-settings', {
        settings: settings
    });
}

// 站点设置相关接口
export interface SiteSettingsConfig {
    siteName: string;
    siteLogo: string;
    siteDescription: string;
    siteKeywords: string;
    siteUrl: string;
    titleTemplate: string;
    defaultDescription: string;
    icpNumber: string;
    timezone: string;
    defaultLanguage: string;
    maintenanceMode: boolean;
    maintenanceMessage: string;
}

export const getSiteSettings = (): Promise<Result<SiteSettingsConfig>> => {
    return axiosInstance.get('api/admin/site-settings');
}

export const saveSiteSettings = (settings: SiteSettingsConfig): Promise<Result<any>> => {
    return axiosInstance.post('api/admin/save-site-settings', {
        settings: settings
    });
}

// Footer管理相关接口
export interface FooterItem {
    name: string;
    url: string;
}

export interface FooterGroup {
    name: string;
    children: FooterItem[];
}

export interface PItem {
    content: string;
}

export interface FooterConfig {
    primary: PItem[];
    list: FooterGroup[];
}

export const getFooterLinks = (): Promise<Result<FooterConfig>> => {
    return axiosInstance.get('api/admin/footer-links');
}

export const saveFooterLinks = (footerConfig: FooterConfig): Promise<Result<any>> => {
    return axiosInstance.post('api/admin/save-footer-links', {
        footerConfig: footerConfig
    });
}

// 赞助商管理相关接口
export const getSponsors = (): Promise<Result<SponsorsConfig>> => {
    return axiosInstance.get('api/admin/sponsors');
}

export const saveSponsors = (sponsorsInfo: SponsorsConfig): Promise<Result<any>> => {
    return axiosInstance.post('api/admin/save-sponsors', {
        sponsorsInfo: sponsorsInfo
    });
}

// 邮件设置相关接口
export interface MailSettingsConfig {
    enableMail: boolean;
    smtpHost: string;
    smtpPort: number;
    useSSL: boolean;
    smtpUsername: string;
    smtpPassword: string;
    fromName: string;
    fromEmail: string;
}

export const getMailSettings = (): Promise<Result<MailSettingsConfig>> => {
    return axiosInstance.get('api/admin/mail-settings');
}

export const saveMailSettings = (settings: MailSettingsConfig): Promise<Result<any>> => {
    return axiosInstance.post('api/admin/save-mail-settings', {
        settings: settings
    });
}

export const testMailConnection = (settings: MailSettingsConfig, testEmail: string): Promise<Result<any>> => {
    return axiosInstance.post('api/admin/test-mail-connection', {
        settings: settings,
        testEmail: testEmail
    });
}
