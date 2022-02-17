export type PingResponse = { latency: number }[];

type PingState =
  | { status: "initial" }
  | {
      status: "started";
      host: string;
      location: string;
      pings: PingResponse[];
    }
  | {
      status: "stopped";
      host: string;
      location: string;
      pings: PingResponse[];
    }
  | { status: "error"; error: Error; host?: string };

type PingAction =
  | { type: "start"; host: string; location: string }
  | { type: "stop" }
  | { type: "ping"; ping: PingResponse }
  | { type: "error"; error: Error };

const pingReducer = (state: PingState, action: PingAction): PingState => {
  switch (action.type) {
    case "start":
      return {
        status: "started",
        host: action.host,
        location: action.location,
        pings: [],
      };
    case "stop":
      return state.status === "started"
        ? { ...state, status: "stopped" }
        : state;
    case "ping":
      return state.status === "started"
        ? { ...state, pings: [...state.pings, action.ping] }
        : state;
    case "error":
      return { ...state, status: "error", error: action.error };
    default:
      return state;
  }
};

export default pingReducer;
