import { z } from "zod";

const bgpRouteSchema = z.object({
  ip: z
    .string({
      required_error: "An IP address is required",
      invalid_type_error: "A valid IP address must be provided",
    })
    .min(1, "An IP address is required"),
  location: z.string(),
});

export default bgpRouteSchema;
