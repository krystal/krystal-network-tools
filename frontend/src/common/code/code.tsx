import {
  Box,
  BoxProps,
  useColorModeValue,
  Button,
  HStack,
  useToast,
} from "@chakra-ui/react";
import { FC } from "react";
import { FaClipboard } from "react-icons/fa";

const Code: FC<BoxProps> = ({ children, ...props }) => {
  const toast = useToast();

  const copyToClipboard = () => {
    const content = children?.toString();
    if (content) {
      navigator.clipboard.writeText(content);
      toast({
        status: "success",
        title: "Copied to your clipboard succesfully",
      });
    }
  };

  return (
    <Box>
      <HStack
        p={2}
        border="1px solid"
        borderColor={useColorModeValue("gray.200", "gray.800")}
        bg={useColorModeValue("white", "gray.700")}
        borderTopRadius="md"
        justify="flex-end"
      >
        <Button size="xs" onClick={copyToClipboard} leftIcon={<FaClipboard />}>
          Copy to clipboard
        </Button>
      </HStack>

      <Box
        as="pre"
        display="block"
        overflowX="auto"
        bg="gray.900"
        color="gray.50"
        p={4}
        fontSize="xs"
        fontFamily="monospace"
        borderBottomRadius="md"
        maxW="100%"
        {...props}
      >
        <code>{children}</code>
      </Box>
    </Box>
  );
};

export default Code;
