import { z } from "zod";

export enum DnsType {
  ANY = "ANY",
  A = "A",
  AAAA = "AAAA",
  MX = "MX",
  NS = "NS",
  PTR = "PTR",
  SOA = "SOA",
  SRV = "SRV",
  TXT = "TXT",
}

const dnsSchema = z.object({
  host: z
    .string({
      required_error: "A hostname or IP address is required",
      invalid_type_error: "A valid hostname or IP address must be provided",
    })
    .min(1, "A hostname or IP address is required"),
  type: z.nativeEnum(DnsType),
  cache: z.boolean(),
  trace: z.boolean(),
});

export default dnsSchema;
