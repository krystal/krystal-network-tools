type Ping = { latency: number }[];

export const useAverageLatency = (pings: Ping[]) => {
  return pings.length
    ? Math.round(
        pings.reduce((total, val) => {
          return total + val[0].latency;
        }, 0) / pings.length
      )
    : 0;
};
