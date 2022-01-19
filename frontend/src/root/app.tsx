import { FC } from "react";
import { Outlet } from "react-router-dom";

import { Box, ChakraProvider, Container } from "@chakra-ui/react";

import theme from "./theme";
import "./app.css";
import Header from "../components/header";

const App: FC = () => {
  return (
    <Box data-testid="app" id="app" h="100%" w="100%">
      <ChakraProvider theme={theme}>
        <Header />
        <Container maxW="container.lg" py={{ base: 6, md: 8 }}>
          <Outlet />
        </Container>
      </ChakraProvider>
    </Box>
  );
};

export default App;
