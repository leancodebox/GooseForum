import {
  BootstrapError,
  type BootstrapErrorProps,
} from "@gooseforum/ui/bootstrap";

const developmentDescription =
  "请先启动 Go 服务，或检查 Vite 的 GOOSE_DEV_ORIGIN 配置。";

export function BrowserBootstrapError({
  description,
  ...props
}: BootstrapErrorProps) {
  return (
    <BootstrapError
      {...props}
      description={
        description ??
        (import.meta.env.DEV ? developmentDescription : undefined)
      }
    />
  );
}
