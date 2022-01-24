export const pingLatencyColor = (latency: number) => {
  if (latency < 80) return "green";
  if (latency > 140) return "red";
  return "orange";
};
