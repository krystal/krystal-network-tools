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
  Heading,
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

  const bgColor = useColorModeValue("gray.50", "gray.700");
  const activeBgColor = useColorModeValue("brand.50", "gray.900");

  return (
    <NavLink to={to}>
      <HStack
        to={to}
        py={2}
        px={3}
        borderRadius="md"
        bg={active ? activeBgColor : undefined}
        _hover={{ bg: active ? activeBgColor : bgColor }}
      >
        <Icon as={icon} w={5} h={5} mr={1} />
        <Text fontWeight="bold">{children}</Text>
      </HStack>
    </NavLink>
  );
};

const NavigationSection: FC<{ title?: string }> = ({ title, children }) => {
  return (
    <Stack as="section" spacing={2}>
      {title && (
        <Heading
          fontSize="xs"
          pl={3}
          textTransform="uppercase"
          pb={2}
          color="gray.500"
        >
          {title}
        </Heading>
      )}

      {children}
    </Stack>
  );
};

const Navigation: FC = () => {
  return (
    <Box as="aside" position="sticky" top="0">
      <Stack as="aside" spacing={8}>
        <NavigationSection>
          <NavigationItem to="/" icon={LogoIcon}>
            tools
          </NavigationItem>
        </NavigationSection>

        <NavigationSection title="Functions">
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
        </NavigationSection>
      </Stack>
    </Box>
  );
};

export default Navigation;
