import { FC } from "react";

import { IconButton, Tooltip, Icon } from "@chakra-ui/react";

import { IconType } from "react-icons";

const TooltipIconButton: FC<{
  label: string;
  onClick?: () => void;
  icon: IconType;
}> = ({ label, onClick, icon }) => {
  return (
    <Tooltip label={label}>
      <IconButton
        size="sm"
        variant="ghost"
        aria-label={label}
        onClick={onClick}
        icon={<Icon as={icon} w={4} h={4} />}
      />
    </Tooltip>
  );
};

export default TooltipIconButton;
