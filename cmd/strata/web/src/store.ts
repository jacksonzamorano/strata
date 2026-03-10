import { createStore } from "solid-js/store";

export interface LogEntry {
  kind: string;
  namespace: string;
  message: string;
  timestamp: string;
}

export interface Task {
  name: string;
  url: string;
}

export interface Component {
  name: string;
  version: string;
  suceeded: boolean;
  path: string;
  error?: string;
}

export interface Trigger {
  id: string;
  name: string;
  date: string;
}

export interface Authorization {
  nickname?: string;
  secret: string;
  source: string;
  created_date: string;
}

export interface PermissionRequest {
  id: string;
  permission: {
    permission: {
      container: string;
      action: string;
      scope: string;
    };
  };
}

export interface SecretRequest {
  id: string;
  namespace: string;
  prompt: string;
}

export interface OauthRequest {
  id: string;
  namespace: string;
  url: string;
  destination: string;
}

export interface AppState {
  connected: boolean;
  kicked: boolean;
  authFailed: boolean;
  logs: LogEntry[];
  tasks: Task[];
  components: Component[];
  triggers: Trigger[];
  authorizations: Authorization[];
  permissionRequest: PermissionRequest | null;
  secretRequest: SecretRequest | null;
  oauthRequest: OauthRequest | null;
}

const [state, setState] = createStore<AppState>({
  connected: false,
  kicked: false,
  authFailed: false,
  logs: [],
  tasks: [],
  components: [],
  triggers: [],
  authorizations: [],
  permissionRequest: null,
  secretRequest: null,
  oauthRequest: null,
});

export { state, setState };
