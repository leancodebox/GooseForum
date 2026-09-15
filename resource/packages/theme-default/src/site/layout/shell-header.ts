import { createContext, useContext, type Dispatch, type SetStateAction } from "react";

export interface ShellHeaderTag {
  id: number | string;
  name: string;
  color?: string;
}

export interface ShellHeaderState {
  title: string;
  tags: ShellHeaderTag[];
  visible: boolean;
}

export const emptyShellHeader: ShellHeaderState = {
  title: "",
  tags: [],
  visible: false,
};

export const ShellHeaderContext = createContext<
  Dispatch<SetStateAction<ShellHeaderState>>
>(() => undefined);

export function useShellHeader() {
  return useContext(ShellHeaderContext);
}
