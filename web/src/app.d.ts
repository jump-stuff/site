// See https://svelte.dev/docs/kit/types#app.d.ts

import type { Event, EventWithLeaderboards, Session, SiteStats } from '$lib/schema';

// for information about these interfaces
declare global {
  namespace App {
    // interface Error {}
    interface Locals {
      session: Session | undefined;
    }
    interface PageData {
      session: Session | undefined;
      player?: Player | undefined;
      players?: Player[] | undefined;
      events?: Event[] | undefined;
      ewl?: EventWithLeaderboards | undefined;
      stats?: SiteStats | undefined;
    }
    // interface PageState {}
    // interface Platform {}
  }

  type OpenAPIError =
    | {
        readonly $schema?: string;
        detail?: string;
        errors?:
          | {
              /** @description Where the error occurred, e.g. 'body.items[3].tags' or 'path.thing-id' */
              location?: string;
              /** @description Error message text */
              message?: string;
              /** @description The value at the given location */
              value?: unknown;
            }[]
          | null;
        instance?: string;
        status?: number;
        title?: string;
        type: string;
      }
    | undefined;
}

export {};
