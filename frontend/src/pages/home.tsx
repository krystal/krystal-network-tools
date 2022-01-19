import { FC } from "react";

import {
  Box,
  Heading,
  HStack,
  Icon,
  SimpleGrid,
  Stack,
  Text,
  useColorModeValue,
} from "@chakra-ui/react";
import { IconType } from "react-icons";
import {
  FcElectricity,
  FcGenealogy,
  FcGlobe,
  FcMindMap,
  FcQuestions,
  FcUndo,
} from "react-icons/fc";
import { Link } from "react-router-dom";

type Page = {
  title: string;
  text: string;
  url: string;
  icon: IconType;
};

const pages: Page[] = [
  {
    title: "Ping",
    text: "Send a ping request to an IPv4 or IPv6 address",
    url: "/ping",
    icon: FcElectricity,
  },
  {
    url: "/traceroute",
    icon: FcMindMap,
    title: "Traceroute",
    text: "Perform a traceroute to a specific address",
  },
  {
    url: "/whois",
    icon: FcQuestions,
    title: "WHOIS",
    text: "Query WHOIS servers for a specific hostname or IP.",
  },
  {
    url: "/dns",
    icon: FcGlobe,
    title: "DNS",
    text: "Find DNS records associated with a hostname.",
  },
  {
    url: "/reverse-dns",
    icon: FcUndo,
    title: "Reverse DNS",
    text: "Find all DNS records associated with a specific IP address.",
  },
  {
    url: "/bgp-route",
    icon: FcGenealogy,
    title: "BGP Route",
    text: "Look up BGP routes for a specific address.",
  },
];

const Home: FC = () => {
  const borderColor = useColorModeValue("gray.200", "gray.700");
  const bgColor = useColorModeValue("white", "gray.700");
  const hoverBgColor = useColorModeValue("brand.400", "brand.300");

  return (
    <SimpleGrid columns={{ base: 1, lg: 2 }} gap={6}>
      {pages.map((page) => (
        <Link to={page.url}>
          <Stack
            py={6}
            px={4}
            align="center"
            justify="center"
            textAlign="center"
            borderRadius="md"
            border="2px solid"
            borderColor={borderColor}
            bg={bgColor}
            height="100%"
            _hover={{ borderColor: hoverBgColor }}
          >
            <Icon w={12} h={12} as={page.icon} mb={4} />
            <Heading size="md">{page.title}</Heading>
            <Text opacity="0.6">{page.text}</Text>
          </Stack>
        </Link>
      ))}
    </SimpleGrid>
  );
};

export default Home;
