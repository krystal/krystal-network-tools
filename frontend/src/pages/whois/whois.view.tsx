import { FC, useState } from "react";

import { Heading, Stack, Text } from "@chakra-ui/react";
import Card from "../../common/card/card";
import WhoisForm from "./whois-form";
import request from "../../api/request";
import endpoint from "../../api/endpoint";

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
          {result.result.split("\n").map((line) => (
            <Text mb={2}>{line}</Text>
          ))}
        </Card>
      )}
    </Stack>
  );
};

export default Whois;
