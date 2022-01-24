import { FC } from "react";
import { Outlet, useLocation } from "react-router-dom";

import {
  Box,
  ChakraProvider,
  Container,
  Grid,
  GridItem,
  HStack,
} from "@chakra-ui/react";

import "./app.css";

import appTheme from "./app-theme";
import Header from "../common/header/header";
import Navigation from "../common/navigation/navigation";

const App: FC = () => {
  const location = useLocation();
  const isHomePage = location.pathname === "/";

  return (
    <Box data-testid="app" id="app" h="100%" w="100%">
      <ChakraProvider theme={appTheme}>
        <Header />

        <Container maxW="container.lg" py={{ base: 6, md: 8 }}>
          <Grid
            templateColumns={{
              base: "1fr",
              md: isHomePage ? "1fr" : "180px 1fr",
            }}
            gap={{ base: 0, md: 16 }}
          >
            {!isHomePage && (
              <GridItem display={{ base: "none", md: "block" }}>
                <Navigation position="sticky" top={6} />
              </GridItem>
            )}

            <GridItem overflowX="hidden">
              <Outlet />
            </GridItem>
          </Grid>
        </Container>
      </ChakraProvider>
    </Box>
  );
};

export default App;
