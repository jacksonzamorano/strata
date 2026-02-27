import type { HostMessageAuthorizationCreated } from "./generated";

export type LogRecord = {
  id: string;
  timestamp: number;
  date: string;
  channel: string;
  kind: string;
  namespace?: string;
  message: string;
  payload?: string;
};

export type RequestState = "in_progress" | "succeeded" | "failed" | "unknown";

export type RequestRecord = {
  id: string;
  taskName: string;
  startedAt: Date | null;
  finishedAt: Date | null;
  durationSeconds?: number;
  state: RequestState;
  sortTs: number;
};

export type TabKey = "overview" | "authorization" | "logs" | "requests";

export type TokenState = HostMessageAuthorizationCreated | null;
export type TokenRecord = HostMessageAuthorizationCreated;

export type RegisteredTask = {
  name: string;
  url: string;
};

export type RegisteredComponent = {
  name: string;
  version: string;
  isHealthy: boolean;
};
