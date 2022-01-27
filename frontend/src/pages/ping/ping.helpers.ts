export const pingLatencyColor = (latency: number) => {
  if (latency === null || isNaN(latency)) return "red";
  if (latency < 70) return "green";
  if (latency > 110) return "orange";
  return "yellow";
};
