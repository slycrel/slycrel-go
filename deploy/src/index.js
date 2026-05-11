import { Container, getContainer } from "@cloudflare/containers";

export class SlycrelServer extends Container {
  defaultPort = 8080;
  // Idle window after the last connection closes before the container sleeps.
  // Short = cheaper but cold-start hits the next player; long = warmer UX.
  sleepAfter = "5m";
}

export default {
  async fetch(request, env) {
    return getContainer(env.SLYCREL).fetch(request);
  },
};
