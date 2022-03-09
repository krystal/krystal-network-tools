import { DnsType } from "./dns.schema";

export const getDnsColor = (type: DnsType) => {
  switch (type) {
    case "A":
      return "blue";
    case "AAAA":
      return "cyan";
    case "CNAME":
      return "orange";
    case "MX":
      return "pink";
    case "NS":
      return "purple";
    case "PTR":
      return "green";
    case "SOA":
      return "red";
    case "SRV":
      return "yellow";
    case "TXT":
      return "teal";
    case "TRACE":
      return "black";
    default:
      return "gray";
  }
};
