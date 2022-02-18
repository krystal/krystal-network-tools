import { FC, useEffect, useReducer } from "react";

import {
  Alert,
  AlertIcon,
  Badge,
  Button,
  Heading,
  HStack,
  Stack,
  Tag,
  Spinner,
} from "@chakra-ui/react";

import endpoint from "../../api/endpoint";
import request from "../../api/request";
import Card from "../../common/card/card";
import pingReducer, { PingResponse } from "./ping.reducer";
import { pingLatencyColor } from "./ping.helpers";
import { useAverageLatency } from "./ping.hooks";
import PingForm from "./ping-form";

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
                <Tag colorScheme={pingLatencyColor(avg)}>{avg}ms</Tag>
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
                  <HStack key={i} justifyContent="space-between">
                    <HStack>
                      <Badge colorScheme="brand">ping</Badge>
                      <Heading size="sm">{state.host}</Heading>
                    </HStack>

                    <Tag colorScheme={pingLatencyColor(ping[0].latency)}>
                      {ping[0].latency}ms
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
