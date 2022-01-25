import { Box, BoxProps, useColorModeValue } from "@chakra-ui/react";
import { FC } from "react";

const Code: FC<BoxProps> = ({ children, ...props }) => {
  return (
    <Box
      as="pre"
      display="inline-block"
      overflowX="auto"
      bg={useColorModeValue("gray.200", "gray.900")}
      color={useColorModeValue("gray.900", "gray.50")}
      py={2}
      px={4}
      fontSize="xs"
      fontFamily="monospace"
      borderRadius="md"
      maxW="100%"
      {...props}
    >
      <code>{children}</code>
    </Box>
  );
};

export default Code;
