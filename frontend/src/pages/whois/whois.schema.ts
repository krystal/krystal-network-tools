import { z } from "zod";

const whoisSchema = z.object({
  host: z.string().min(1),
});

export default whoisSchema;
