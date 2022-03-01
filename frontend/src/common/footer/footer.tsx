import { FC } from "react";

import {
  Box,
  Container,
  Tag,
  useColorModeValue,
  TagLeftIcon,
  TagLabel,
} from "@chakra-ui/react";
import LogoIcon from "../icons/logo-icon";
import { Link } from "react-router-dom";

const Footer: FC = () => {
  return (
    <Box
      as="footer"
      borderTop={useColorModeValue("3px solid", "1px solid")}
      borderTopColor={useColorModeValue("gray.200", "gray.800")}
      transition="border 0.2s ease"
      py={6}
      mt={8}
    >
      <Container maxW="container.lg" textAlign="center">
        <Tag
          as={Link}
          to="https://krystal.uk/"
          colorScheme="brand"
          size="lg"
          _hover={{ textDecoration: "underline" }}
        >
          <TagLeftIcon as={LogoIcon} />
          <TagLabel>Powered by Krystal</TagLabel>
        </Tag>
      </Container>
    </Box>
  );
};

export default Footer;
