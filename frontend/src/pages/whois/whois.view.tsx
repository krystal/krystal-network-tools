import { FC, useState } from "react";

import { Heading, Stack, Text } from "@chakra-ui/react";
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
      <Heading size="lg">WHOIS</Heading>

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
          <Code copyText={result.result}>
            {result.result.split("\n").map((line, index) => {
              const skip = line.includes(">>>") || line.includes("--");
              if (line.match(/^(?![%#])[a-zA-Z0-9\s\-_/]{1,40}:/) && !skip) {
                const [val, ...rest] = line.split(":");
                return (
                  <Text key={index}>
                    <Text as="span" color="gray.500">
                      {val}:
                    </Text>
                    {rest.join(":")}
                  </Text>
                );
              } else {
                return <Text key={index}>{line}</Text>;
              }
            })}
          </Code>
        </Card>
      )}
    </Stack>
  );
};

export default Whois;
