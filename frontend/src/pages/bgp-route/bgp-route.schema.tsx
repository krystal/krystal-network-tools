import { z } from "zod";
import { ServerLocation } from "../../api/get-api-url";

const bgpRouteSchema = z.object({
  ip: z
    .string({
      required_error: "An IP address is required",
      invalid_type_error: "A valid IP address must be provided",
    })
    .min(1, "An IP address is required"),
  location: z.nativeEnum(ServerLocation),
});

export default bgpRouteSchema;
