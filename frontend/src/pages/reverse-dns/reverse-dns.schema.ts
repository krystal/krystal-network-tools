import { z } from "zod";

const reverseDnsSchema = z.object({
  ip: z.string().min(1),
});

export default reverseDnsSchema;
