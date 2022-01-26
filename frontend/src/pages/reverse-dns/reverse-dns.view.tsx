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
  const [result, setResult] = useState<ReverseDnsResponse | null>(null);

  return (
    <Stack spacing={6}>
      <Heading size="md">Reverse DNS</Heading>

      <Card>
        <ReverseDnsForm
          disabled={false}
          onSubmit={async ({ ip }) => {
            const result = await request<ReverseDnsResponse>(
              endpoint("/rdns/:ip", { ip })
            );
            setResult(result);
          }}
        />
      </Card>

      {result !== null && (
        <Card>
          <Code>{JSON.stringify(result, null, 2)}</Code>
        </Card>
      )}
    </Stack>
  );
};

export default ReverseDns;
