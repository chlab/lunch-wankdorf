export async function dispatchWorkflow(
  env,
  workflow,
  fetchImpl = globalThis.fetch,
) {
  const url = `https://api.github.com/repos/${env.GH_OWNER}/${env.GH_REPO}/actions/workflows/${workflow}/dispatches`;

  const response = await fetchImpl(url, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${env.GH_DISPATCH_TOKEN}`,
      Accept: "application/vnd.github+json",
      "X-GitHub-Api-Version": "2022-11-28",
      "User-Agent": "lunch-wankdorf-scheduler",
    },
    body: JSON.stringify({ ref: env.GH_REF }),
  });

  if (!response.ok) {
    const body = await response.text();
    throw new Error(`dispatch ${workflow}: ${response.status} ${body}`);
  }
}
