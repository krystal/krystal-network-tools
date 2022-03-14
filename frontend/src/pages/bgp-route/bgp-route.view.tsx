import { FC, Fragment, useState } from "react";

import { Heading, Stack, Table, Td, Th, Tr } from "@chakra-ui/react";
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
          onSubmit={async ({ ip, location }) => {
            const result = await request<BgpRouteResponse>(
              endpoint("/bgp/:ip", { ip }),
              { location }
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

              <Table size="sm">
                <Tr>
                  <Th>Next hop</Th>
                  <Td colSpan={2}>
                    <Code>{item.next_hop}</Code>
                  </Td>
                </Tr>
                <Tr>
                  <Th>As path</Th>
                  <Td colSpan={2}>
                    <Code>{item.as_path?.join(", ")}</Code>
                  </Td>
                </Tr>
                <Tr>
                  <Th>Community</Th>
                  <Td colSpan={2}>
                    <Code>{item.community?.join(" - ")}</Code>
                  </Td>
                </Tr>
                {!!item.large_community && (
                  <Tr>
                    <Th>Large community</Th>
                    <Td colSpan={2}>
                      <Code>{item.large_community?.join(" - ")}</Code>
                    </Td>
                  </Tr>
                )}
                <Tr>
                  <Th>Local pref</Th>
                  <Td colSpan={2}>
                    <Code>{item.local_pref}</Code>
                  </Td>
                </Tr>
              </Table>
            </Card>
          ))}
        </Fragment>
      )}
    </Stack>
  );
};

export default BgpRoute;
