// Dispatches each workflow, carrying on past a failure so one broken job does
// not hide the others, then throws so the Cron invocation is recorded as failed.
export async function dispatchAll(workflows, env, { dispatch, log = console }) {
  const failures = [];

  for (const workflow of workflows) {
    if (env.DRY_RUN === "true") {
      log.log(`dry run: would dispatch ${workflow}`);
      continue;
    }

    try {
      await dispatch(workflow);
      log.log(`dispatched ${workflow}`);
    } catch (error) {
      log.error(error.message);
      failures.push(error.message);
    }
  }

  if (failures.length > 0) throw new Error(failures.join("; "));
}
