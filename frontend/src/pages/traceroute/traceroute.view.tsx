import { FC, useEffect, useReducer } from "react";

import {
  Alert,
  AlertIcon,
  Button,
  Heading,
  HStack,
  Stack,
  Text,
  Spinner,
} from "@chakra-ui/react";

import endpoint from "../../api/endpoint";
import request from "../../api/request";
import Card from "../../common/card/card";
import tracerouteReducer, { TracerouteResponse } from "./traceroute.reducer";
import TracerouteForm from "./traceroute-form";
import Code from "../../common/code/code";
import TraceroutePings from "./traceroute-pings";

const Traceroute: FC = () => {
  const [state, dispatch] = useReducer(tracerouteReducer, {
    status: "initial",
  });

  const makeRequest = (host: string, hop: number, location: string) => {
    return request<TracerouteResponse>(
      endpoint("/traceroute/:host", { host, hop }),
      { location }
    )
      .then((data) => {
        dispatch({ type: "response", response: data });
      })
      .catch((error) => dispatch({ type: "error", error }));
  };

  useEffect(() => {
    if (state.status === "started") {
      makeRequest(state.host, 1, state.location);
    }
  }, [state.status]); // eslint-disable-line

  useEffect(() => {
    if (state.status === "started" && state.responses.length) {
      const last = state.responses[state.responses.length - 1];
      if (last.destination_ip === last.traceroute[0]?.ip_address) {
        dispatch({ type: "stop" });
      } else {
        makeRequest(
          last.destination_ip,
          state.responses.length + 1,
          state.location
        );
      }
    }
  }, [state]);

  return (
    <Stack spacing={6}>
      <Heading size="lg">Traceroute</Heading>

      <Card>
        <TracerouteForm
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
                {state.status === "started" && <Spinner size="xs" />}
                <Heading size="sm">{state.host}</Heading>
              </HStack>
              <Button
                size="sm"
                colorScheme="red"
                isDisabled={state.status !== "started"}
                onClick={() => dispatch({ type: "stop" })}
              >
                Stop
              </Button>
            </HStack>

            {state.status === "error" && (
              <Alert status="error" variant="solid" borderRadius="sm">
                <AlertIcon />
                Error getting traceroute for {state.host || "host"}
              </Alert>
            )}

            {state.status !== "error" && (
              <Stack>
                {state.responses.map(({ traceroute }, i) => (
                  <HStack key={i} justifyContent="space-between">
                    <Stack>
                      {traceroute[0] ? (
                        <Code>
                          {traceroute[0].ip_address}
                          {traceroute[0].rdns && (
                            <Text as="span" opacity="0.5">
                              {` (${traceroute[0].rdns})`}
                            </Text>
                          )}
                        </Code>
                      ) : (
                        <Code>*</Code>
                      )}
                    </Stack>

                    <TraceroutePings pings={traceroute[0]?.pings || []} />
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

export default Traceroute;
