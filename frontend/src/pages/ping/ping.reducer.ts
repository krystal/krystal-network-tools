import { ServerLocation } from "../../api/get-api-url";

type PingState =
  | { status: "initial" }
  | { status: "started"; host: string; location: ServerLocation }
  | { status: "stopped"; host: string; location: ServerLocation }
  | { status: "error"; error: Error; host?: string };

type PingAction =
  | { type: "start"; host: string; location: ServerLocation }
  | { type: "stop" }
  | { type: "error"; error: Error };

const pingReducer = (state: PingState, action: PingAction): PingState => {
  switch (action.type) {
    case "start":
      return {
        status: "started",
        host: action.host,
        location: action.location,
      };
    case "stop":
      return state.status === "started"
        ? { status: "stopped", host: state.host, location: state.location }
        : state;
    case "error":
      return { ...state, status: "error", error: action.error };
    default:
      return state;
  }
};

export default pingReducer;
