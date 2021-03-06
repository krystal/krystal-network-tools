import { FC, Fragment, useState } from "react";

import { Box, Heading, Stack, Tag, Text } from "@chakra-ui/react";
import Card from "../../common/card/card";
import DnsForm from "./dns-form";
import request from "../../api/request";
import endpoint from "../../api/endpoint";
import { DnsType } from "./dns.schema";
import DnsTable from "./dns-table";
import { getDnsColor } from "./dns.helpers";

type DnsRecord = {
  type: DnsType;
  ttl: number;
  priority?: number;
  name: string;
  dnsServer: string;
  value:
    | string
    | string[]
    | {
        expire: number;
        mbox: string;
        minttl: number;
        ns: string;
        refresh: number;
        retry: number;
        serial: number;
      };
};

export type DnsResponse = {
  [key in DnsType]: {
    server: string;
    records: DnsRecord[];
  }[];
};

const Dns: FC = () => {
  const [result, setResult] = useState<DnsResponse | null>(null);

  const empty =
    result && !Object.values(result).find((record) => !!record.length);

  return (
    <Stack spacing={6}>
      <Heading size="lg">DNS</Heading>

      <Card>
        <DnsForm
          disabled={false}
          onSubmit={async ({ host, type, trace }) => {
            const result = await request<DnsResponse>(
              endpoint("/dns/:type/:host", { host, type, trace })
            );
            setResult(result);
          }}
        />
      </Card>

      <Stack spacing={6}>
        {result !== null && empty && (
          <Card>
            <Text color="gray.500">No DNS records were found.</Text>
          </Card>
        )}

        {result !== null &&
          !empty &&
          (Object.keys(result) as DnsType[]).map((type) => {
            const record = result[type];

            if (!record.length) return null;

            return (
              <Fragment key={type}>
                <Box>
                  <Tag colorScheme={getDnsColor(type)} size="lg">
                    {type}
                  </Tag>
                </Box>
                <Card overflowX="auto">
                  <DnsTable record={record} />
                </Card>
              </Fragment>
            );
          })}
      </Stack>
    </Stack>
  );
};

export default Dns;
