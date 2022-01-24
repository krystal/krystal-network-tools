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
  host: z.string().min(1),
  type: z.nativeEnum(DnsType),
});

export default dnsSchema;
