export type TracerouteResponse = {
  destination_ip: string;
  traceroute: {
    pings: number[];
    rdns: string;
    ip_address: string;
  }[];
};

type TracerouteState =
  | { status: "initial" }
  | {
      status: "started";
      host: string;
      location: string;
      responses: TracerouteResponse[];
    }
  | {
      status: "stopped";
      host: string;
      location: string;
      responses: TracerouteResponse[];
    }
  | { status: "error"; error: Error; host?: string };

type TracerouteAction =
  | { type: "start"; host: string; location: string }
  | { type: "stop" }
  | { type: "response"; response: TracerouteResponse }
  | { type: "error"; error: Error };

const tracerouteReducer = (
  state: TracerouteState,
  action: TracerouteAction
): TracerouteState => {
  switch (action.type) {
    case "start":
      return {
        status: "started",
        host: action.host,
        location: action.location,
        responses: [],
      };
    case "stop":
      return state.status === "started"
        ? { ...state, status: "stopped" }
        : state;
    case "response":
      return state.status === "started"
        ? { ...state, responses: [...state.responses, action.response] }
        : state;
    case "error":
      return { ...state, status: "error", error: action.error };
    default:
      return state;
  }
};

export default tracerouteReducer;
