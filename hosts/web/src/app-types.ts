import type { HostMessageAuthorizationCreated } from "./generated";

export type LogRecord = {
  id: string;
  date: string;
  channel: string;
  kind: string;
  namespace?: string;
  message: string;
  payload?: string;
};

export type TabKey = "overview" | "authorization" | "logs";

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
