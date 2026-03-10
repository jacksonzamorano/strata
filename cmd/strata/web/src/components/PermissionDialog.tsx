import { state, setState } from "../store";
import { send } from "../ws";

export default function PermissionDialog() {
  const req = () => state.permissionRequest!;
  const perm = () => req().permission.permission;

  function respond(approve: boolean) {
    send("permissionResponse", req().id, { approve });
    setState("permissionRequest", null);
  }

  return (
    <div class="modal-backdrop">
      <div class="modal animate-in">
        <div class="modal-header">
          <span class="modal-icon">⚐</span>
          <span class="modal-title">Permission Request</span>
        </div>
        <div class="modal-body">
          <div class="modal-row">
            <span class="modal-label">container</span>
            <span class="modal-value">{perm().container}</span>
          </div>
          <div class="modal-row">
            <span class="modal-label">action</span>
            <span class="modal-value">{perm().action}</span>
          </div>
          <div class="modal-row">
            <span class="modal-label">scope</span>
            <span class="modal-value">{perm().scope}</span>
          </div>
        </div>
        <div class="modal-footer">
          <button
            onClick={() => respond(false)}
            class="btn btn-ghost"
            style="flex: 1;"
          >
            Deny
          </button>
          <button
            onClick={() => respond(true)}
            class="btn btn-primary"
            style="flex: 1;"
          >
            Approve
          </button>
        </div>
      </div>
    </div>
  );
}
