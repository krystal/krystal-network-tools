import { FC } from "react";

import { Box, BoxProps, useColorModeValue } from "@chakra-ui/react";

const Card: FC<BoxProps> = ({ children, ...props }) => {
  const borderColor = useColorModeValue("gray.200", "gray.800");
  const borderWidth = useColorModeValue("2px", "1px");
  const bgColor = useColorModeValue("white", "gray.700");

  return (
    <Box
      p={6}
      borderRadius="md"
      border={`${borderWidth} solid`}
      borderColor={borderColor}
      bg={bgColor}
      w="100%"
      overflow="hidden"
      {...props}
    >
      {children}
    </Box>
  );
};

export default Card;