export interface Result<T> {
    code: 0 | 1; // 0 成功 1 失败
    result: T;
    msg: string;
}

export interface PageData<T> {
    list: T[];
    page: number
    size: number
    total: number
}

export interface QueryList<T> {
    list: T[];
}

// 类型定义
export interface ArticleData {
    id: number
    content: string
    title: string
    categoryId: number[]
    type: number
}

export interface ArticleListItem {
    id: number,
    title: string,
    username: string,
    createTime: string,
    lastUpdateTime: string,
    viewCount: number,
    commentCount: number
    category: string,
    categories: string[]
    typeStr?: string
}


export interface EnumInfoResponse {
    code: number;
    result: {
        category: NameLabel[];
        type: NameLabel[];
    };
}

export interface NameLabel {
    name: string;
    value: number;
}


export interface Payload {
    title: string,
    content: string,
    actorId: number,
    actorName: string,
    articleId: number,
    articleTitle: string
    commentId?: number
}

export interface Notifications {
    id: number,
    userId: number,
    payload: Payload,
    eventType: string,
    isRead: boolean,
    readAt: string | null,
    createdAt: string,
    updatedAt: string
}

export interface ExternalInformationItem {
    link: string
}

export interface ExternalInformation {
    github: ExternalInformationItem,
    weibo: ExternalInformationItem,
    bilibili: ExternalInformationItem,
    twitter: ExternalInformationItem,
    linkedIn: ExternalInformationItem,
    zhihu: ExternalInformationItem,
}

export interface AuthorInfoStatistics {
    userId: number,
    articleCount: number,
    replyCount: number,
    followerCount: number,
    followingCount: number,
    likeReceivedCount: number,
    likeGivenCount: number,
    collectionCount: number,
}

// 定义用户表单接口
export interface UserInfo {
    userId?: number
    avatarUrl: string
    username: string
    nickname: string
    email: string
    bio: string
    website: string
    websiteName: string
    signature: string,
    externalInformation: ExternalInformation
    authorInfoStatistics?: AuthorInfoStatistics
}

// OAuth相关接口
export interface OAuthBinding {
    bound: boolean
    provider?: string
    createdAt?: string
    updatedAt?: string
}

export interface OAuthBindings {
    github?: OAuthBinding
    google?: OAuthBinding

    [key: string]: OAuthBinding | undefined
}
