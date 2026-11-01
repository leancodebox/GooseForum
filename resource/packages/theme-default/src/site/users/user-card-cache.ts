import type { GooseSiteApi, UserCardPayload } from "@gooseforum/client";

const cardCache = new Map<number, UserCardPayload>();
const pendingCards = new Map<number, Promise<UserCardPayload>>();

export function getCachedUserCard(userId: number) {
  return cardCache.get(userId);
}

export function invalidateUserCard(userId: number) {
  cardCache.delete(userId);
  pendingCards.delete(userId);
}

export function loadUserCard(api: GooseSiteApi, userId: number) {
  const cached = cardCache.get(userId);
  if (cached) return Promise.resolve(cached);
  const pending = pendingCards.get(userId);
  if (pending) return pending;
  const request = api.users.card(userId)
    .then((card) => {
      if (pendingCards.get(userId) === request) cardCache.set(userId, card);
      return card;
    })
    .finally(() => {
      if (pendingCards.get(userId) === request) pendingCards.delete(userId);
    });
  pendingCards.set(userId, request);
  return request;
}
