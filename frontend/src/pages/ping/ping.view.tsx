import { FC, useEffect, useReducer, useState } from "react";

import {
  Alert,
  AlertIcon,
  Badge,
  Button,
  Center,
  Heading,
  HStack,
  Stack,
  Tag,
  Spinner,
} from "@chakra-ui/react";

import endpoint from "../../api/endpoint";
import request from "../../api/request";
import Card from "../../common/card/card";
import pingReducer from "./ping.reducer";
import { pingLatencyColor } from "./ping.helpers";
import { useAverageLatency } from "./ping.hooks";
import PingForm from "./ping-form";

type PingResponse = { latency: number }[];

const Ping: FC = () => {
  const [state, dispatch] = useReducer(pingReducer, { status: "initial" });
  const [pings, setPings] = useState<PingResponse[]>([]);

  const ping = (host: string) => {
    return request<PingResponse>(endpoint("/ping/:host", { host }))
      .then((data) => {
        setPings((pings) => [...pings, data]);
      })
      .catch((error) => dispatch({ type: "error", error }));
  };

  useEffect(() => {
    if (state.status === "started") {
      setPings([]);
      const timer = window.setInterval(() => ping(state.host), 500);
      return () => {
        clearTimeout(timer);
      };
    }
  }, [state]);

  useEffect(() => {
    if (pings.length >= 10) {
      dispatch({ type: "stop" });
    }
  }, [pings]);

  const avg = useAverageLatency(pings);

  return (
    <Stack spacing={6}>
      <Heading size="md">Ping</Heading>

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
              <Tag colorScheme={pingLatencyColor(avg)}>{avg}ms</Tag>
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
              <Alert status="error">
                <AlertIcon />
                Could not connect to {state.host || "host"}
              </Alert>
            )}

            {state.status !== "error" &&
              (pings.length ? (
                <Stack>
                  {pings.map((ping, i) => (
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
              ) : (
                <Center p={1}>
                  <Spinner size="lg" />
                </Center>
              ))}
          </Stack>
        </Card>
      )}
    </Stack>
  );
};

export default Ping;
