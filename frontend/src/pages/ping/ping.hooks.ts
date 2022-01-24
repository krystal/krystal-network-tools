type Ping = { latency: number }[];

export const useAverageLatency = (pings: Ping[]) => {
  return pings.length
    ? Math.round(
        pings.reduce((total, val) => {
          if (isNaN(val[0].latency)) return total;
          return total + val[0].latency;
        }, 0) / pings.length
      )
    : 0;
};
