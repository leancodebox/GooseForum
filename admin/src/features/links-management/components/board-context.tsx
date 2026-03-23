import { type LinkGroup } from '../data/schema';

export type BoardContextValue = {
  getGroups: () => LinkGroup[];
  reorderGroup: (args: { startIndex: number; finishIndex: number }) => void;
  reorderLink: (args: { groupIdx: number; startIndex: number; finishIndex: number }) => void;
  moveLink: (args: {
    startGroupIdx: number;
    finishGroupIdx: number;
    itemIndexInStartGroup: number;
    itemIndexInFinishGroup?: number;
  }) => void;
  instanceId: symbol;
  registerGroup: (groupId: string, entry: { element: HTMLElement }) => () => void;
  registerLink: (linkId: string, entry: { element: HTMLElement; actionMenuTrigger: HTMLElement }) => () => void;
};

import { createContext, useContext } from 'react';

export const BoardContext = createContext<BoardContextValue | null>(null);

export function useBoardContext() {
  const context = useContext(BoardContext);
  if (!context) {
    throw new Error('useBoardContext must be used within a BoardProvider');
  }
  return context;
}
