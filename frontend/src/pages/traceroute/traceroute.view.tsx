import { FC, useState } from "react";

import { Heading, Stack } from "@chakra-ui/react";
import Card from "../../common/card/card";
import TracerouteForm from "./traceroute-form";
import request from "../../api/request";
import endpoint from "../../api/endpoint";
import Code from "../../common/code/code";

type TracerouteResponse = {
  pings: number[];
  rdns: string;
  ip_address: string;
}[];

const Traceroute: FC = () => {
  const [result, setResult] = useState<TracerouteResponse[] | null>(null);

  return (
    <Stack spacing={6}>
      <Heading size="md">Traceroute</Heading>

      <Card>
        <TracerouteForm
          disabled={false}
          onSubmit={async ({ host }) => {
            setResult([]);

            let hop = 1;

            while (hop <= 64) {
              const res = await request<TracerouteResponse>(
                endpoint("/traceroute/:host", { host, hop })
              );

              setResult((state) => (state ? [...state, res] : [res]));

              if (res[0] && res[0].ip_address === host) {
                break;
              } else {
                hop++;
              }
            }
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

export default Traceroute;
