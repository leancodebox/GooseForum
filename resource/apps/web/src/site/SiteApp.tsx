import {
  SiteApp as SharedSiteApp,
  type SiteAppProps,
} from "@gooseforum/theme-default/app/site-app";

export type { SiteAppProps };

export function SiteApp(props: SiteAppProps) {
  return (
    <SharedSiteApp
      {...props}
      errorDescription={
        props.errorDescription ??
        (import.meta.env.DEV
          ? "请先启动 Go 服务，或检查 Vite 的 GOOSE_DEV_ORIGIN 配置。"
          : undefined)
      }
    />
  );
}
