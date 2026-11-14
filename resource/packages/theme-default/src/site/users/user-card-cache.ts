import type { GooseSiteApi, UserCardPayload } from "@gooseforum/client";

const cardCache = new Map<number, UserCardPayload>();
const pendingCards = new Map<number, Promise<UserCardPayload>>();
const maxCachedCards = 128;

export function getCachedUserCard(userId: number) {
  const cached = cardCache.get(userId);
  if (cached) {
    cardCache.delete(userId);
    cardCache.set(userId, cached);
  }
  return cached;
}

export function invalidateUserCard(userId: number) {
  cardCache.delete(userId);
  pendingCards.delete(userId);
}

export function loadUserCard(api: GooseSiteApi, userId: number) {
  const cached = getCachedUserCard(userId);
  if (cached) return Promise.resolve(cached);
  const pending = pendingCards.get(userId);
  if (pending) return pending;
  const request = api.users.card(userId)
    .then((card) => {
      if (pendingCards.get(userId) === request) {
        cardCache.delete(userId);
        cardCache.set(userId, card);
        while (cardCache.size > maxCachedCards) {
          const oldest = cardCache.keys().next().value;
          if (oldest === undefined) break;
          cardCache.delete(oldest);
        }
      }
      return card;
    })
    .finally(() => {
      if (pendingCards.get(userId) === request) pendingCards.delete(userId);
    });
  pendingCards.set(userId, request);
  return request;
}
