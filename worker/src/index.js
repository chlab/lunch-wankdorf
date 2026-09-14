import { dispatchAll } from "./dispatch-all.js";
import { dispatchWorkflow } from "./github.js";
import { withRetry } from "./retry.js";
import { dueJobs } from "./schedule.js";

export default {
  async scheduled(controller, env) {
    const workflows = dueJobs(new Date(controller.scheduledTime));

    await dispatchAll(workflows, env, {
      dispatch: (workflow) =>
        withRetry(() => dispatchWorkflow(env, workflow)),
    });
  },
};
