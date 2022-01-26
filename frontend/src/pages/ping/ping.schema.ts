import { z } from "zod";
import { ServerLocation } from "../../api/get-api-url";

const pingSchema = z.object({
  host: z
    .string({
      required_error: "A hostname or IP address is required",
      invalid_type_error: "A valid hostname or IP address must be provided",
    })
    .min(1, "A hostname or IP address is required"),
  location: z.nativeEnum(ServerLocation),
});

export default pingSchema;
