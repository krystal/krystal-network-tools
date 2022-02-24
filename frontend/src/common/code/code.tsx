import { Box, BoxProps, useColorModeValue } from "@chakra-ui/react";
import { FC } from "react";
import { FaClipboard } from "react-icons/fa";

const Code: FC<BoxProps> = ({ children, ...props }) => {
  const copyToClipboard = () => {
    const c = children?.toString();
    if (c) navigator.clipboard.writeText(c);
  };

  return <>
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
    <a onClick={copyToClipboard}>
      <FaClipboard></FaClipboard>
    </a>
  </>;
};

export default Code;
