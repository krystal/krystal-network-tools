import { FC, useEffect } from "react";

import {
  Drawer,
  DrawerOverlay,
  DrawerContent,
  DrawerBody,
  DrawerCloseButton,
  useColorModeValue,
} from "@chakra-ui/react";

import Navigation from "../navigation/navigation";
import { useLocation } from "react-router-dom";

type HeaderDrawerProps = {
  isOpen: boolean;
  onClose: () => void;
};

const HeaderDrawer: FC<HeaderDrawerProps> = ({ isOpen, onClose }) => {
  const bg = useColorModeValue("white", "gray.800");
  const location = useLocation();

  useEffect(() => {
    onClose();
  }, [location]); //eslint-disable-line

  return (
    <Drawer isOpen={isOpen} onClose={onClose} placement="right">
      <DrawerOverlay />
      <DrawerContent bg={bg} p={4}>
        <DrawerCloseButton onClose={onClose} zIndex={2} />
        <DrawerBody pt={7}>
          <Navigation showHomeLink />
        </DrawerBody>
      </DrawerContent>
    </Drawer>
  );
};

export default HeaderDrawer;
