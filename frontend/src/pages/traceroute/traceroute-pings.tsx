import { FC } from "react";

import {
  Popover,
  SimpleGrid,
  Tag,
  PopoverTrigger,
  PopoverContent,
  PopoverArrow,
  PopoverBody,
} from "@chakra-ui/react";

import { pingLatencyColor } from "../ping/ping.helpers";
import { useAverageLatency } from "../ping/ping.hooks";

type TraceroutePingsProps = {
  pings: number[];
};

const TraceroutePings: FC<TraceroutePingsProps> = ({ pings }) => {
  const avg = useAverageLatency(pings.map((ping) => [{ latency: ping }]));

  return (
    <Popover trigger="hover">
      <PopoverTrigger>
        <Tag
          cursor="default"
          colorScheme={pingLatencyColor(avg)}
          fontFamily={!pings.length ? "monospace" : "sans-serif"}
        >
          {!!pings.length ? `${avg}ms` : "*"}
        </Tag>
      </PopoverTrigger>
      {!!pings.length && (
        <PopoverContent w="auto">
          <PopoverArrow />
          <PopoverBody borderRadius="md">
            <SimpleGrid columns={3} gap={2}>
              {pings.map((ping, i) => (
                <Tag key={i} colorScheme={pingLatencyColor(ping)}>
                  {ping >= 0 ? `${ping}ms` : ""}
                </Tag>
              ))}
            </SimpleGrid>
          </PopoverBody>
        </PopoverContent>
      )}
    </Popover>
  );
};

export default TraceroutePings;
