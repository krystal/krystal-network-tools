export const pingLatencyColor = (latency: number) => {
  if (latency < 70) return "green";
  if (latency > 110) return "red";
  return "orange";
};
