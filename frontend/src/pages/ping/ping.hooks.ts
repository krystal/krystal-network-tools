import { useMemo } from "react";

type Ping = { latency: number }[];

export const useAverageLatency = (pings: Ping[]) => {
  return useMemo(() => {
    const pingCount = pings.filter((val) => !isNaN(val[0].latency)).length;

    if (pingCount < 1) return null;

    return pings.length
      ? Number(
          (
            pings.reduce((total, val) => {
              if (isNaN(val[0].latency)) return total;
              return total + val[0].latency;
            }, 0) / pings.filter((val) => !isNaN(val[0].latency)).length
          ).toFixed(3)
        )
      : null;
  }, [pings]);
};
