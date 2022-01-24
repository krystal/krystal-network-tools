import { FC, useState } from "react";

import { Heading, Stack } from "@chakra-ui/react";
import Card from "../../common/card/card";
import DnsForm from "./dns-form";
import request from "../../api/request";
import endpoint from "../../api/endpoint";
import { DnsType } from "./dns.schema";

type SoaValue = {
  expire: number;
  mbox: string;
  minttl: number;
  ns: string;
  refresh: number;
  retry: number;
  serial: number;
};

type DnsResponse = {
  [key in DnsType]: {
    type: DnsType;
    ttl: number;
    priority?: number;
    name: string;
    value: string | string[] | SoaValue;
  }[];
};

const Dns: FC = () => {
  const [result, setResult] = useState<DnsResponse | null>(null);

  return (
    <Stack spacing={6}>
      <Heading size="md">DNS</Heading>

      <Card>
        <DnsForm
          disabled={false}
          onSubmit={async ({ host, type }) => {
            const result = await request<DnsResponse>(
              endpoint("/dns/:type/:host", { host, type })
            );
            setResult(result);
          }}
        />
      </Card>

      {result !== null && (
        <Card>
          <pre>{JSON.stringify(result, null, 2)}</pre>
        </Card>
      )}
    </Stack>
  );
};

export default Dns;
