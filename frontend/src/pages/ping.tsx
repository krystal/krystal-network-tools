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
import { z } from "zod";
import endpoint from "../helpers/endpoint";
import Form from "../components/form";
import request from "../helpers/request";
import TextField from "../components/text-field";
import Card from "../components/card";

type PingResponse = { latency: number }[];

type PingState =
  | { status: "initial" }
  | { status: "started"; host: string }
  | { status: "stopped"; host: string }
  | { status: "error"; error: Error; host?: string };

type PingAction =
  | { type: "start"; host: string }
  | { type: "stop" }
  | { type: "error"; error: Error };

function pingReducer(state: PingState, action: PingAction): PingState {
  switch (action.type) {
    case "start":
      return { status: "started", host: action.host };
    case "stop":
      return state.status === "started"
        ? { status: "stopped", host: state.host }
        : state;
    case "error":
      return { ...state, status: "error", error: action.error };
    default:
      return state;
  }
}

const pingSchema = z.object({
  host: z.string().min(1),
});

const pingLatencyColor = (latency: number) => {
  if (latency < 80) return "green";
  if (latency > 140) return "red";
  return "orange";
};

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

  const avg = pings.length
    ? Math.round(
        pings.reduce((total, val) => {
          return total + val[0].latency;
        }, 0) / pings.length
      )
    : 0;

  return (
    <Stack spacing={6}>
      <Heading size="md">Ping</Heading>

      <Card>
        <Form
          schema={pingSchema}
          initialValues={{ host: "" }}
          onSubmit={({ host }) => {
            dispatch({ type: "start", host });
          }}
          render={(form) => (
            <Stack align="flex-end" spacing={3}>
              <TextField
                name="host"
                isDisabled={state.status === "started"}
                variant="filled"
                label="Hostname or IP address"
                placeholder="Enter the address that you want to ping"
              />
              <Button
                colorScheme="green"
                type="submit"
                px={6}
                isDisabled={state.status === "started"}
              >
                Start Pinging
              </Button>
            </Stack>
          )}
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

            {pings.length ? (
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
              state.status !== "error" && (
                <Center p={4}>
                  <Spinner size="lg" />
                </Center>
              )
            )}
          </Stack>
        </Card>
      )}
    </Stack>
  );
};

export default Ping;
