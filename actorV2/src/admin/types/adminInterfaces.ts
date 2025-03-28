
export interface Result<T> {
  code: number;
  result: T;
  message: string;
}


export interface Role {
  name: string;
  value: number;
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
