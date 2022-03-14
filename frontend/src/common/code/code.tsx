import {
  Box,
  BoxProps,
  IconButton,
  DarkMode,
  Icon,
  HStack,
  Tooltip,
} from "@chakra-ui/react";
import { FC } from "react";
import { IoCopy } from "react-icons/io5";
import { useClipboard } from "../../hooks/use-clipboard";

type CodeProps = {
  copyText?: string;
};

const Code: FC<CodeProps & BoxProps> = ({
  children,
  copyText = "",
  ...props
}) => {
  const childText = typeof children === "string" ? children : "";
  const textToCopy = copyText || childText;

  const showActions = !!textToCopy;

  const copyToClipboard = useClipboard(textToCopy);

  return (
    <Box display="inline-flex" maxW="100%" textAlign="left">
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
          p={2}
          borderRightRadius="md"
          align="flex-start"
          flex="0 0 auto"
        >
          <DarkMode>
            <Tooltip label="Copy to clipboard">
              <IconButton
                size="xs"
                fontFamily="sans-serif"
                onClick={copyToClipboard}
                icon={<Icon h={3} w={3} color="white" as={IoCopy} />}
                aria-label="copy"
                h={4}
              />
            </Tooltip>
          </DarkMode>
        </HStack>
      )}
    </Box>
  );
};

export default Code;
