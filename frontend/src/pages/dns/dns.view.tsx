import { FC, Fragment, useState } from "react";

import { Box, Heading, Stack, Tag, Text } from "@chakra-ui/react";
import Card from "../../common/card/card";
import DnsForm from "./dns-form";
import request from "../../api/request";
import endpoint from "../../api/endpoint";
import { DnsType } from "./dns.schema";
import DnsTable from "./dns-table";
import { getDnsColor } from "./dns.helpers";

export type DnsResponse = {
  [key in DnsType]: {
    type: DnsType;
    ttl: number;
    priority?: number;
    name: string;
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

      <Stack spacing={6}>
        {result !== null &&
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
                  {record.length > 0 ? (
                    <DnsTable record={record} />
                  ) : (
                    <Text color="gray.500">There are no {type} records.</Text>
                  )}
                </Card>
              </Fragment>
            );
          })}
      </Stack>
    </Stack>
  );
};

export default Dns;
