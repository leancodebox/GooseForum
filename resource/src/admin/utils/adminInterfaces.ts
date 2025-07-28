export interface Result<T> {
    code: 0 | 1;
    result: T;
    msg: string;
}
export interface PageData<T> {
    list: T[];
    page: number
    size: number
    total: number
}


export interface Role {
    name: string;
    value: number;
}

export interface Label {
    name: string | null
    label: string | null
    value: number | null
}

export interface User {
    userId: number;
    username: string;
    avatarUrl: string;
    email: string;
    roleId: number;
    status: number;
    validate: number;
    prestige: number;
    roleList: Role[];
    createTime: string;
}

export interface Articles {
    id: number,
    title: string
    type: number
    userId: number
    username: string
    articleStatus: number,
    processStatus: number,
    createdAt: string
    updatedAt: string
}

export interface AdminArticlesItem {
    id: number,
    title: string
    description: string
    type: number
    userId: number
    username: string
    userAvatarUrl: string,
    articleStatus: number,
    processStatus: number,
    viewCount: number
    replyCount: number
    likeCount: number
    createdAt: string
    updatedAt: string
}

export interface Category {
    id: number
    category: string,
    desc: string,
    sort: number
    status: number
}

export interface Permissions {
    name: string;
    id: number;
}

export interface UserRole {
    roleId: number,
    roleName: string,
    effective: number,
    permissions: Permissions[]
    createTime: string
}


export interface ApplySheet {
    id: number,
    userId: number,
    applyUserInfo: string,
    type: number,
    status: number,
    title: string,
    content: string
    createTime: string,
    updateTime: string
}


export interface FriendLinksGroup {
    name: string,
    links: LinkItem[]
}

export interface LinkItem {
    name: string
    desc: string
    url: string
    logoUrl: string
    status: 0 | 1 | any
}

export interface SponsorItem {
    name: string
    logo: string
    info: string
    url: string
    tag: string[]
}

export interface UserSponsor {
  name: string
  logo: string
  amount: string
  time: string
}

export interface SponsorsConfig {
    sponsors: {
        level0: SponsorItem[]
        level1: SponsorItem[]
        level2: SponsorItem[]
        level3: SponsorItem[]
    }
    users: UserSponsor[]
}

export interface FooterItem {
    id: number
    title: string
    url: string
    sort: number
    status: number
    createTime: string
}

export interface FooterGroup {
    name: string
    items: FooterItem[]
}
