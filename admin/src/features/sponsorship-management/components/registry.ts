export function createRegistry() {
  const levels = new Map<string, { element: HTMLElement }>();
  const sponsors = new Map<string, { element: HTMLElement; actionMenuTrigger: HTMLElement }>();
  const users = new Map<string, { element: HTMLElement; actionMenuTrigger: HTMLElement }>();

  return {
    registerLevel(levelId: string, entry: { element: HTMLElement }) {
      levels.set(levelId, entry);
      return () => {
        levels.delete(levelId);
      };
    },
    registerSponsor(sponsorId: string, entry: { element: HTMLElement; actionMenuTrigger: HTMLElement }) {
      sponsors.set(sponsorId, entry);
      return () => {
        sponsors.delete(sponsorId);
      };
    },
    registerUser(userId: string, entry: { element: HTMLElement; actionMenuTrigger: HTMLElement }) {
      users.set(userId, entry);
      return () => {
        users.delete(userId);
      };
    },
    getLevel(levelId: string) {
      const entry = levels.get(levelId);
      if (!entry) throw new Error(`Level ${levelId} not found`);
      return entry;
    },
    getSponsor(sponsorId: string) {
      const entry = sponsors.get(sponsorId);
      if (!entry) throw new Error(`Sponsor ${sponsorId} not found`);
      return entry;
    },
    getUser(userId: string) {
      const entry = users.get(userId);
      if (!entry) throw new Error(`User ${userId} not found`);
      return entry;
    },
  };
}
