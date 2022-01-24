import type { FC } from "react";

import {
  FcInfo,
  FcElectricity,
  FcQuestions,
  FcSearch,
  FcMindMap,
  FcFeedIn,
} from "react-icons/fc";

import { Stack, Box } from "@chakra-ui/react";

import LogoIcon from "../icons/logo-icon";
import NavigationItem from "./navigation-item";

const Navigation: FC<{ showHomeLink?: boolean }> = ({ showHomeLink }) => {
  return (
    <Box as="aside" position="sticky" top={6}>
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
