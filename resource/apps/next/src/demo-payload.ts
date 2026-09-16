import type { AnyPagePayload } from "@gooseforum/client";

export function createDemoPage(url: string): AnyPagePayload {
  const sort = new URL(url, "http://next.local").searchParams.get("sort") || "latest";
  return {
    component: "home.index",
    props: {
      sort,
      tabs: [
        { key: "latest", label: "Latest", url: "/", active: sort === "latest" },
        { key: "hot", label: "Hot", url: "/?sort=hot", active: sort === "hot" },
      ],
      topics: [],
      pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" },
      announcement: {
        enabled: true,
        html: "<p><strong>Standalone Next host</strong> · This page is rendered from the shared GooseForum payload contract without a Go process.</p>",
        publishedAt: "2026-09-16 00:00:00",
      },
    },
    layout: {
      site: {
        name: "GooseForum Next",
        description: "Standalone Next.js host adapter",
        logo: "",
        favicon: "",
        brandType: "default",
        brandText: "",
        brandImage: "",
      },
      viewer: {
        id: 0,
        username: "",
        email: "",
        avatarUrl: "",
        isAuthenticated: false,
        canAccessAdmin: false,
        isModerator: false,
        requiresEmailVerification: false,
        adminPermissions: [],
      },
      header: [],
      sidebar: {
        activeKey: "topics",
        categories: [],
      },
      footer: { links: [], primary: [] },
      unread: { notifications: false, messages: false },
      theme: {
        enabled: false,
        current: "gf-light",
        themeColor: "#fbfdff",
      },
    },
    meta: {
      title: "GooseForum Next",
      description: "Standalone Next.js host adapter for GooseForum themes.",
      robots: "noindex,nofollow",
    },
    url,
    version: "1.0",
  };
}
