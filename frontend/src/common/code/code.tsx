import {
  Box,
  BoxProps,
  useToast,
  IconButton,
  DarkMode,
  HStack,
} from "@chakra-ui/react";
import { FC } from "react";
import { FaClipboard } from "react-icons/fa";

type CodeProps = {
  copyText?: string;
};

const Code: FC<CodeProps & BoxProps> = ({
  children,
  copyText = "",
  ...props
}) => {
  const toast = useToast();

  const childText = typeof children === "string" ? children : "";
  const textToCopy = copyText || childText;

  const showActions = !!textToCopy;

  const copyToClipboard = () => {
    navigator.clipboard.writeText(textToCopy);
    toast({
      status: "success",
      title: "Copied to your clipboard succesfully",
    });
  };

  return (
    <Box display="inline-flex" maxW="100%">
      <Box
        as="pre"
        display="inline-block"
        overflowX="auto"
        bg="gray.900"
        color="gray.50"
        p={4}
        py={2}
        pr={showActions ? 2 : 4}
        fontSize="xs"
        fontFamily="monospace"
        borderLeftRadius="md"
        borderRightRadius={showActions ? 0 : "md"}
        maxW="100%"
        flex="1 1 100%"
        {...props}
      >
        <code>{children}</code>
      </Box>

      {showActions && (
        <HStack
          bg="gray.900"
          px={2}
          pt={2}
          borderRightRadius="md"
          align="flex-start"
          flex="0 0 auto"
        >
          <DarkMode>
            <IconButton
              size="xs"
              fontFamily="sans-serif"
              onClick={copyToClipboard}
              icon={<FaClipboard />}
              aria-label="copy"
              h={4}
            />
          </DarkMode>
        </HStack>
      )}
    </Box>
  );
};

export default Code;
