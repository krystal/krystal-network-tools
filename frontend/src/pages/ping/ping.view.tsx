import { FC, useEffect, useReducer } from "react";

import {
  Alert,
  AlertIcon,
  Box,
  Button,
  Heading,
  HStack,
  Stack,
  Tag,
  Spinner,
  Text,
} from "@chakra-ui/react";

import endpoint from "../../api/endpoint";
import request from "../../api/request";
import Card from "../../common/card/card";
import pingReducer, { PingResponse } from "./ping.reducer";
import { pingLatencyColor } from "./ping.helpers";
import { useAverageLatency } from "./ping.hooks";
import PingForm from "./ping-form";
import Code from "../../common/code/code";

const Ping: FC = () => {
  const [state, dispatch] = useReducer(pingReducer, { status: "initial" });

  const ping = (host: string, location: string) => {
    return request<PingResponse>(endpoint("/ping/:host", { host }), {
      location,
    })
      .then((data) => {
        dispatch({ type: "ping", ping: data });
      })
      .catch((error) => dispatch({ type: "error", error }));
  };

  useEffect(() => {
    if (state.status === "started") {
      if (state.pings.length >= 10) {
        dispatch({ type: "stop" });
      } else {
        const timer = window.setInterval(
          () => ping(state.host, state.location),
          500
        );
        return () => {
          clearTimeout(timer);
        };
      }
    }
  }, [state]);

  const avg = useAverageLatency(
    state.status === "started" || state.status === "stopped" ? state.pings : []
  );

  return (
    <Stack spacing={6}>
      <Heading size="lg">Ping</Heading>

      <Card>
        <PingForm
          disabled={state.status === "started"}
          onSubmit={({ host, location }) => {
            dispatch({ type: "start", host, location });
          }}
        />
      </Card>

      {state.status !== "initial" && (
        <Card>
          <Stack spacing={6}>
            <HStack justify="space-between">
              <HStack>
                {state.status === "started" && <Spinner size="sm" />}
                {(state.status === "started" || state.status === "stopped") &&
                  state.pings.length && (
                    <Tag colorScheme={pingLatencyColor(avg)}>
                      {avg ? `${avg}ms` : "error"}
                    </Tag>
                  )}
              </HStack>
              <Button
                size="sm"
                colorScheme="red"
                isDisabled={state.status !== "started"}
                onClick={() => dispatch({ type: "stop" })}
              >
                Stop pinging
              </Button>
            </HStack>

            {state.status === "error" && (
              <Alert status="error" variant="solid" borderRadius="sm">
                <AlertIcon />
                Could not connect to {state.host || "host"}
              </Alert>
            )}

            {state.status !== "error" && (
              <Stack>
                {state.pings.map((ping, i) => (
                  <HStack key={i} justifyContent="space-between" maxW="100%">
                    <Box flexShrink={1} minW={0}>
                      <Code>
                        {ping[0].ip_address}
                        {ping[0].hostname && (
                          <Text as="span" opacity="0.5">
                            {` (${ping[0].hostname})`}
                          </Text>
                        )}
                      </Code>
                    </Box>

                    <Tag
                      flex="0 0 auto"
                      colorScheme={pingLatencyColor(ping[0].latency)}
                    >
                      {ping[0].latency === null || isNaN(ping[0].latency)
                        ? "error"
                        : `${ping[0].latency}ms`}
                    </Tag>
                  </HStack>
                ))}
              </Stack>
            )}
          </Stack>
        </Card>
      )}
    </Stack>
  );
};

export default Ping;
