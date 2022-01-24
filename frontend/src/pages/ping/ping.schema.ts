import { z } from "zod";
import { ServerLocation } from "../../api/get-api-url";

const pingSchema = z.object({
  host: z.string().min(1),
  location: z.nativeEnum(ServerLocation),
});

export default pingSchema;
