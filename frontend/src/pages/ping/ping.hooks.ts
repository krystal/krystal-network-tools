import { useMemo } from "react";

type Ping = { latency: number }[];

export const useAverageLatency = (pings: Ping[]) => {
  return useMemo(() => {
    return pings.length
      ? Math.round(
          pings.reduce((total, val) => {
            if (isNaN(val[0].latency)) return total;
            return total + val[0].latency;
          }, 0) / pings.filter((val) => !isNaN(val[0].latency)).length
        )
      : 0;
  }, [pings]);
};