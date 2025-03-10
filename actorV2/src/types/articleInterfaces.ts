export interface ArticleInfo {
  id: number;
  articleContent: string;
  articleTitle: string;
  categoryId: number[];
  type: number;
}

export interface ArticleResponse {
  code: number;
  result: ArticleInfo;
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