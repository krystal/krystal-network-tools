import { Box } from "@chakra-ui/react";
import { FC } from "react";

const Code: FC = ({ children }) => {
  return (
    <Box
      as="pre"
      overflowX="auto"
      bg="black"
      color="white"
      p={4}
      borderRadius="md"
    >
      <code>{children}</code>
    </Box>
  );
};

export default Code;
