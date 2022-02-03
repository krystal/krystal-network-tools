import { FC, useState } from "react";

import { Heading, Stack } from "@chakra-ui/react";
import Card from "../../common/card/card";
import ReverseDnsForm from "./reverse-dns-form";
import request from "../../api/request";
import endpoint from "../../api/endpoint";
import Code from "../../common/code/code";

type ReverseDnsResponse = {
  hostname: string;
};

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
        <Card>
          <Heading size="sm" mb={4}>
            {ip}
          </Heading>
          <Code fontSize="lg" w="100%">
            {result.hostname}
          </Code>
        </Card>
      )}
    </Stack>
  );
};

export default ReverseDns;
