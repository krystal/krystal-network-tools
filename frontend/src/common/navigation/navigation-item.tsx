import type { FC } from "react";

import { NavLink, useMatch } from "react-router-dom";

import { HStack, Text, useColorModeValue, Icon } from "@chakra-ui/react";

type NavigationItemProps = {
  to: string;
  icon: any;
};

const NavigationItem: FC<NavigationItemProps> = ({ children, icon, to }) => {
  const active = useMatch(to);

  const activeColor = useColorModeValue("brand.500", "brand.200");

  return (
    <NavLink to={to}>
      <HStack
        to={to}
        py={2}
        px={3}
        borderRadius="md"
        color={active ? activeColor : undefined}
        _hover={{ color: activeColor }}
      >
        <Icon as={icon} w={5} h={5} mr={1} />
        <Text pr={5} fontWeight="bold">
          {children}
        </Text>
      </HStack>
    </NavLink>
  );
};

export default NavigationItem;
