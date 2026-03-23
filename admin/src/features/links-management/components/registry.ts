export function createRegistry() {
  const groups = new Map<string, { element: HTMLElement }>();
  const links = new Map<string, { element: HTMLElement; actionMenuTrigger: HTMLElement }>();

  return {
    registerGroup(groupId: string, entry: { element: HTMLElement }) {
      groups.set(groupId, entry);
      return () => {
        groups.delete(groupId);
      };
    },
    registerLink(linkId: string, entry: { element: HTMLElement; actionMenuTrigger: HTMLElement }) {
      links.set(linkId, entry);
      return () => {
        links.delete(linkId);
      };
    },
    getGroup(groupId: string) {
      const entry = groups.get(groupId);
      if (!entry) throw new Error(`Group ${groupId} not found`);
      return entry;
    },
    getLink(linkId: string) {
      const entry = links.get(linkId);
      if (!entry) throw new Error(`Link ${linkId} not found`);
      return entry;
    },
  };
}
