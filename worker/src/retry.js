const defaultSleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

export async function withRetry(
  fn,
  { attempts = 3, baseDelayMs = 1000, sleep = defaultSleep } = {},
) {
  for (let attempt = 1; ; attempt++) {
    try {
      return await fn();
    } catch (error) {
      if (attempt >= attempts) throw error;
      await sleep(baseDelayMs * 2 ** (attempt - 1));
    }
  }
}
