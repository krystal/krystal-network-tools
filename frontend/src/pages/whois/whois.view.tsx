import { FC, useState } from "react";

import { Heading, Stack } from "@chakra-ui/react";
import Card from "../../common/card/card";
import WhoisForm from "./whois-form";
import request from "../../api/request";
import endpoint from "../../api/endpoint";
import Code from "../../common/code/code";

type WhoisResponse = {
  result: string;
};

const Whois: FC = () => {
  const [result, setResult] = useState<WhoisResponse | null>(null);

  return (
    <Stack spacing={6}>
      <Heading size="md">WHOIS</Heading>

      <Card>
        <WhoisForm
          disabled={false}
          onSubmit={async ({ host }) => {
            const result = await request<WhoisResponse>(
              endpoint("/whois/:host", { host })
            );
            setResult(result);
          }}
        />
      </Card>

      {result !== null && (
        <Card>
          <Code>{result.result}</Code>
        </Card>
      )}
    </Stack>
  );
};

export default Whois;