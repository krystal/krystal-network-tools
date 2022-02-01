import { FC, Fragment, useState } from "react";

import { Heading, Stack } from "@chakra-ui/react";
import Card from "../../common/card/card";
import BgpRouteForm from "./bgp-route-form";
import request from "../../api/request";
import endpoint from "../../api/endpoint";
import Code from "../../common/code/code";

type BgpRouteResponse = {
  prefix: string;
  as_path: string[];
  local_pref: number;
  next_hop: string;
  community: string[];
  large_community: string[];
}[];

const BgpRoute: FC = () => {
  const [ip, setIp] = useState("");
  const [result, setResult] = useState<BgpRouteResponse | null>(null);

  return (
    <Stack spacing={6}>
      <Heading size="md">BGP Route</Heading>

      <Card>
        <BgpRouteForm
          disabled={false}
          onSubmit={async ({ ip }) => {
            const result = await request<BgpRouteResponse>(
              endpoint("/bgp/:ip", { ip })
            );
            setIp(ip);
            setResult(result);
          }}
        />
      </Card>

      {result !== null && (
        <Fragment>
          {result.map((item, index) => (
            <Card key={index}>
              <Heading size="sm" mb={4}>
                {item.prefix}
              </Heading>
              <Code>{JSON.stringify(item, null, 2)}</Code>
            </Card>
          ))}
        </Fragment>
      )}
    </Stack>
  );
};

export default BgpRoute;
