export interface Result<T> {
    code: 0 | 1;
    result: T;
    message: string;
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

export interface ArticleInfo {
    id: number;
    articleContent: string;
    articleTitle: string;
    categoryId: number[];
    type: number;
}

export interface ArticleListItem {
    id: number,
    title: string,
    createTime: string,
    lastUpdateTime: string,
    viewCount: number,
    commentCount: number
    category: string,
    categories: string[]
    typeStr?:string
}

export interface ArticleResponse {
    code: number;
    result: ArticleInfo;
    message: string;
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
}

export interface Notifications {
    id: number,
    userId: number,
    payload: Payload,
    eventType: string,
    isRead: false,
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
    userId:number,
    articleCount:number,
    replyCount:number,
    followerCount:number,
    followingCount:number,
    likeReceivedCount:number,
    likeGivenCount:number,
    collectionCount:number,
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
    authorInfoStatistics ?: AuthorInfoStatistics
}
