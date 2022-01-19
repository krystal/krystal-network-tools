import { FC } from "react";

import {
  Heading,
  Icon,
  SimpleGrid,
  Stack,
  Text,
  useColorModeValue,
} from "@chakra-ui/react";
import { IconType } from "react-icons";
import {
  FcElectricity,
  FcFeedIn,
  FcSearch,
  FcMindMap,
  FcQuestions,
  FcInfo,
} from "react-icons/fc";
import { Link } from "react-router-dom";
import Card from "../components/card";

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
    icon: FcSearch,
    title: "DNS",
    text: "Find DNS records associated with a hostname.",
  },
  {
    url: "/reverse-dns",
    icon: FcInfo,
    title: "Reverse DNS",
    text: "Find all DNS records associated with a specific IP address.",
  },
  {
    url: "/bgp-route",
    icon: FcFeedIn,
    title: "BGP Route",
    text: "Look up BGP routes for a specific address.",
  },
];

const Home: FC = () => {
  const textColor = useColorModeValue("gray.500", "gray.300");
  const hoverBgColor = useColorModeValue("brand.400", "brand.300");

  return (
    <SimpleGrid columns={{ base: 1, lg: 2 }} gap={6}>
      {pages.map((page) => (
        <Link to={page.url} key={page.url}>
          <Card
            height="100%"
            _hover={{ borderColor: hoverBgColor, color: hoverBgColor }}
          >
            <Stack align="center" justify="center" textAlign="center">
              <Icon w={10} h={10} as={page.icon} mb={4} />
              <Heading size="md">{page.title}</Heading>
              <Text color={textColor}>{page.text}</Text>
            </Stack>
          </Card>
        </Link>
      ))}
    </SimpleGrid>
  );
};

export default Home;
