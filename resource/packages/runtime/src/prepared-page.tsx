import {
  createElement,
  lazy,
  type ComponentType,
  type ComponentProps,
} from "react";

const pageLoaders = new Map<string, () => Promise<unknown>>();

export function preparedPage<Component extends ComponentType<any>>(
  name: string,
  loader: () => Promise<{ default: Component }>,
): ComponentType<ComponentProps<Component>> {
  let component: Component | undefined;
  let pending: ReturnType<typeof loader> | undefined;
  const load = () => {
    if (!pending) {
      pending = loader()
        .then((module) => {
          component = module.default;
          return module;
        })
        .catch((error) => {
          pending = undefined;
          throw error;
        });
    }
    return pending;
  };
  pageLoaders.set(name, load);
  const LazyPage = lazy(load);
  return function PreparedPage(props: ComponentProps<Component>) {
    return component
      ? createElement(component, props)
      : createElement(LazyPage, props);
  };
}

export async function prepareGoosePage(name: string) {
  await pageLoaders.get(name)?.();
}
