import { FC, useState } from "react";

import {
  Box,
  Container,
  Heading,
  HStack,
  useColorMode,
  useColorModeValue,
  Tag,
  TagLeftIcon,
  TagLabel,
  Grid,
  GridItem,
  Tooltip,
} from "@chakra-ui/react";

import { FaBars, FaMoon } from "react-icons/fa";
import { BiNetworkChart } from "react-icons/bi";
import { BsSunFill } from "react-icons/bs";

import useQuery from "../../api/use-query";
import LogoIcon from "../icons/logo-icon";
import { Link } from "react-router-dom";
import HeaderDrawer from "./header-drawer";
import TooltipIconButton from "../tooltip-icon-button/tooltip-icon-button";

const Header: FC = () => {
  const { data } = useQuery<{ ip: string }>("/ip");
  const { toggleColorMode } = useColorMode();
  const [menuIsOpen, setMenuIsOpen] = useState(false);

  return (
    <Box
      as="header"
      borderTop="4px solid"
      borderColor={useColorModeValue("brand.400", "brand.300")}
      bg={useColorModeValue("white", "gray.700")}
      borderBottom="1px solid"
      borderBottomColor={useColorModeValue("gray.200", "gray.800")}
      transition="border 0.2s ease"
      py={2}
    >
      <HeaderDrawer isOpen={menuIsOpen} onClose={() => setMenuIsOpen(false)} />

      <Container maxW="container.lg">
        <Grid
          templateColumns={{ base: "auto 1fr 1fr", md: "1fr auto 1fr" }}
          alignItems="center"
          gap={3}
        >
          <GridItem>
            <HStack>
              <TooltipIconButton
                icon={FaBars}
                onClick={() => setMenuIsOpen(true)}
                label="Open menu"
              />
              <TooltipIconButton
                icon={useColorModeValue(FaMoon, BsSunFill)}
                onClick={toggleColorMode}
                label={useColorModeValue(
                  "Switch to dark mode",
                  "Switch to light mode"
                )}
              />
            </HStack>
          </GridItem>

          <GridItem
            justify={{ base: "flex-start", md: "center" }}
            order={{ base: -1, md: 0 }}
          >
            <HStack as={Link} to="/">
              <LogoIcon />
              <Heading display={{ base: "none", md: "block" }} size="md">
                tools
              </Heading>
            </HStack>
          </GridItem>

          <GridItem>
            <HStack justify="flex-end">
              {data?.ip && (
                <Tooltip label="Your current IP address">
                  <Tag variant="subtle" colorScheme="brand">
                    <TagLeftIcon boxSize="12px" as={BiNetworkChart} />
                    <TagLabel
                      maxWidth={{
                        base: "120px",
                        sm: "320px",
                        md: "200px",
                        lg: "320px",
                      }}
                      isTruncated
                    >
                      {data.ip}
                    </TagLabel>
                  </Tag>
                </Tooltip>
              )}
            </HStack>
          </GridItem>
        </Grid>
      </Container>
    </Box>
  );
};

export default Header;
