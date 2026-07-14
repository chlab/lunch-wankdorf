/**
 * Deterministic randomness, seeded by date. The restaurant order and the daily
 * recommendation should feel random but stay put for the whole day: reshuffling
 * on every recompute (a filter toggle, a day switch) just looks like a glitch.
 */

// mulberry32
const createRandom = (seed) => {
  let state = seed >>> 0;
  return () => {
    state = (state + 0x6d2b79f5) >>> 0;
    let result = Math.imul(state ^ (state >>> 15), 1 | state);
    result = (result + Math.imul(result ^ (result >>> 7), 61 | result)) ^ result;
    return ((result ^ (result >>> 14)) >>> 0) / 4294967296;
  };
};

/** Same date -> same seed, e.g. 20260714. */
export const dateSeed = (date) =>
  date.getFullYear() * 10000 + (date.getMonth() + 1) * 100 + date.getDate();

/** Fisher-Yates, so every order is equally likely (unlike sort(() => Math.random() - 0.5)). */
export function shuffle(items, seed) {
  const random = createRandom(seed);
  const shuffled = [...items];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
}

export function pick(items, seed) {
  return items[Math.floor(createRandom(seed)() * items.length)];
}
