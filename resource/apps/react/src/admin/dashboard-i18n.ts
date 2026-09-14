import type { AuthLocale } from "@gooseforum/react/i18n/auth";
const en = {
  title: "Dashboard",
  description: "Site activity, traffic, and runtime version.",
  users: "Users",
  topics: "Topics",
  posts: "Replies",
  links: "Links",
  registrations: "Registrations",
  traffic: "Traffic overview",
  trafficHint: "Daily registrations, topics, and replies.",
  last7Days: "Last 7 days",
  last30Days: "Last 30 days",
  last90Days: "Last 90 days",
  selectRange: "Select a date range",
  apply: "Apply",
  cancel: "Cancel",
  version: "Version",
  releases: "Project releases",
  viewAll: "View all",
  loading: "Loading…",
  empty: "No data",
  release: "Release",
  snapshot: "Snapshot",
  development: "Development",
  custom: "Custom",
} as const;
const zh: Record<keyof typeof en, string> = {
  title: "仪表盘",
  description: "站点活动、流量和运行版本。",
  users: "用户",
  topics: "主题",
  posts: "回复",
  links: "友情链接",
  registrations: "注册",
  traffic: "流量概览",
  trafficHint: "每日注册、主题和回复趋势。",
  last7Days: "近 7 天",
  last30Days: "近 30 天",
  last90Days: "近 90 天",
  selectRange: "请选择日期范围",
  apply: "应用",
  cancel: "取消",
  version: "版本",
  releases: "项目版本",
  viewAll: "查看全部",
  loading: "正在加载…",
  empty: "暂无数据",
  release: "正式版",
  snapshot: "快照版",
  development: "开发版",
  custom: "自定义",
};
const resources = {
  zh,
  en,
  ja: { ...en, title: "ダッシュボード" },
  it: { ...en, title: "Dashboard" },
} as const;
export type DashboardTextKey = keyof typeof en;
export function createDashboardText(l: AuthLocale) {
  const d = resources[l];
  return (k: DashboardTextKey) => d[k];
}
