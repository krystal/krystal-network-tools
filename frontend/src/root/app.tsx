import { FC } from "react";
import { Outlet, useLocation } from "react-router-dom";

import { Box, ChakraProvider, Container, HStack } from "@chakra-ui/react";

import theme from "./theme";
import "./app.css";
import Header from "../components/header";
import Navigation from "../components/navigation";

const App: FC = () => {
  const location = useLocation();

  return (
    <Box data-testid="app" id="app" h="100%" w="100%">
      <ChakraProvider theme={theme}>
        <Header />
        <Container maxW="container.lg" py={{ base: 6, md: 8 }}>
          <HStack align="flex-start" spacing={{ base: 0, md: 16 }}>
            {location.pathname !== "/" && (
              <Box display={{ base: "none", md: "block" }} flexShrink={0}>
                <Navigation />
              </Box>
            )}

            <Box w="100%">
              <Outlet />
            </Box>
          </HStack>
        </Container>
      </ChakraProvider>
    </Box>
  );
};

export default App;
