import { type SponsorsConfig, Sponsors } from '../data/schema';

export type BoardContextValue = {
  getConfig: () => SponsorsConfig;
  reorderSponsor: (args: { sourceLevel: keyof Sponsors; destLevel: keyof Sponsors; startIndex: number; finishIndex: number }) => void;
  reorderUser: (args: { startIndex: number; finishIndex: number }) => void;
  instanceId: symbol;
  registerLevel: (levelId: string, entry: { element: HTMLElement }) => () => void;
  registerSponsor: (sponsorId: string, entry: { element: HTMLElement; actionMenuTrigger: HTMLElement }) => () => void;
  registerUser: (userId: string, entry: { element: HTMLElement; actionMenuTrigger: HTMLElement }) => () => void;
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
