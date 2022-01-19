import type { FC } from "react";

import { NavLink, useMatch } from "react-router-dom";
import LogoIcon from "./logo-icon";

import {
  FcInfo,
  FcElectricity,
  FcQuestions,
  FcSearch,
  FcMindMap,
  FcFeedIn,
} from "react-icons/fc";

import {
  Stack,
  HStack,
  Text,
  Box,
  useColorModeValue,
  Icon,
} from "@chakra-ui/react";

type NavigationItemProps = {
  to: string;
  icon: any;
};

const NavigationItem: FC<NavigationItemProps> = ({ children, icon, to }) => {
  const active = useMatch(to);

  const bgColor = useColorModeValue("gray.50", "gray.900");
  const activeColor = useColorModeValue("brand.500", "brand.200");

  return (
    <NavLink to={to}>
      <HStack
        to={to}
        py={2}
        px={3}
        borderRadius="md"
        bg={active ? bgColor : undefined}
        color={active ? activeColor : undefined}
        _hover={{ bg: active ? bgColor : bgColor }}
      >
        <Icon as={icon} w={5} h={5} mr={1} />
        <Text pr={5} fontWeight="bold">
          {children}
        </Text>
      </HStack>
    </NavLink>
  );
};

const Navigation: FC<{ showHomeLink?: boolean }> = ({ showHomeLink }) => {
  return (
    <Box as="aside" position="sticky" top="0">
      <Stack as="aside" spacing={2}>
        {showHomeLink && (
          <NavigationItem to="/" icon={LogoIcon}>
            tools
          </NavigationItem>
        )}

        <NavigationItem to="/ping" icon={FcElectricity}>
          Ping
        </NavigationItem>
        <NavigationItem to="/traceroute" icon={FcMindMap}>
          Traceroute
        </NavigationItem>
        <NavigationItem to="/whois" icon={FcQuestions}>
          WHOIS
        </NavigationItem>
        <NavigationItem to="/dns" icon={FcSearch}>
          DNS
        </NavigationItem>
        <NavigationItem to="/reverse-dns" icon={FcInfo}>
          Reverse DNS
        </NavigationItem>
        <NavigationItem to="/bgp-route" icon={FcFeedIn}>
          BGP Route
        </NavigationItem>
      </Stack>
    </Box>
  );
};

export default Navigation;
