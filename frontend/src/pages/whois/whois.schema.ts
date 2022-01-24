import { z } from "zod";

const whoisSchema = z.object({
  host: z
    .string({
      required_error: "A hostname or IP address is required",
      invalid_type_error: "A valid hostname or IP address must be provided",
    })
    .min(1, "A hostname or IP address is required"),
});

export default whoisSchema;
