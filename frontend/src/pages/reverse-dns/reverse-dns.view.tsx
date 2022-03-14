import { FC, Fragment, useState } from "react";

import { Box, Heading, Stack, Tag } from "@chakra-ui/react";
import Card from "../../common/card/card";
import ReverseDnsForm from "./reverse-dns-form";
import request from "../../api/request";
import endpoint from "../../api/endpoint";
import Code from "../../common/code/code";
import { DnsType } from "../dns/dns.schema";
import { DnsResponse } from "../dns/dns.view";
import DnsTable from "../dns/dns-table";

type ReverseDnsResponse =
  | { trace: DnsResponse[DnsType] }
  | { hostname: string };

const ReverseDns: FC = () => {
  const [ip, setIp] = useState("");
  const [result, setResult] = useState<ReverseDnsResponse | null>(null);

  return (
    <Stack spacing={6}>
      <Heading size="lg">Reverse DNS</Heading>

      <Card>
        <ReverseDnsForm
          disabled={false}
          onSubmit={async ({ ip, trace }) => {
            const result = await request<ReverseDnsResponse>(
              endpoint("/rdns/:ip", { ip, trace })
            );
            setIp(ip);
            setResult(result);
          }}
        />
      </Card>

      {result !== null && (
        <Fragment>
          <Box>
            <Tag colorScheme="brand" size="lg">
              {ip}
            </Tag>
          </Box>

          {"trace" in result ? (
            <Card overflowX="auto">
              <DnsTable record={result.trace} />
            </Card>
          ) : (
            <Card>
              <Code fontSize="lg">{result.hostname}</Code>
            </Card>
          )}
        </Fragment>
      )}
    </Stack>
  );
};

export default ReverseDns;
