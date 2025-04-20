export interface Result<T> {
    code: 0 | 1;
    result: T;
    message: string;
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
    email: string;
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

export interface Category {
    id: number
    category: string,
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
