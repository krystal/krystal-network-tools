import { FC, Fragment, useState } from "react";

import { Heading, SimpleGrid, Stack, Text } from "@chakra-ui/react";
import Card from "../../common/card/card";
import BgpRouteForm from "./bgp-route-form";
import request from "../../api/request";
import endpoint from "../../api/endpoint";
import Code from "../../common/code/code";

export type BgpRouteResponse = {
  prefix: string;
  as_path: string[];
  local_pref: number;
  next_hop: string;
  community: string[][];
  large_community: string[][];
}[];

const BgpRoute: FC = () => {
  const [result, setResult] = useState<BgpRouteResponse | null>(null);

  return (
    <Stack spacing={6}>
      <Heading size="lg">BGP Route</Heading>

      <Card>
        <BgpRouteForm
          disabled={false}
          onSubmit={async ({ ip }) => {
            const result = await request<BgpRouteResponse>(
              endpoint("/bgp/:ip", { ip })
            );
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
              <Code display="block" py={4}>
                <Stack>
                  <SimpleGrid columns={2}>
                    <Heading size="xs" opacity={0.5} fontFamily="monospace">
                      Next hop
                    </Heading>
                    <Text>{item.next_hop}</Text>
                  </SimpleGrid>
                  <SimpleGrid columns={2}>
                    <Heading size="xs" opacity={0.5} fontFamily="monospace">
                      As path
                    </Heading>
                    <Text>{item.as_path?.join(", ")}</Text>
                  </SimpleGrid>
                  <SimpleGrid columns={2}>
                    <Heading size="xs" opacity={0.5} fontFamily="monospace">
                      Community
                    </Heading>
                    <Text>{item.community?.join(" - ")}</Text>
                  </SimpleGrid>
                  <SimpleGrid columns={2}>
                    <Heading size="xs" opacity={0.5} fontFamily="monospace">
                      Large community
                    </Heading>
                    <Text>{item.large_community?.join(" - ")}</Text>
                  </SimpleGrid>
                  <SimpleGrid columns={2}>
                    <Heading size="xs" opacity={0.5} fontFamily="monospace">
                      Local pref
                    </Heading>
                    <Text>{item.local_pref}</Text>
                  </SimpleGrid>
                </Stack>
              </Code>
            </Card>
          ))}
        </Fragment>
      )}
    </Stack>
  );
};

export default BgpRoute;
