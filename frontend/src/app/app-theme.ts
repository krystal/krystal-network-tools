import { extendTheme } from "@chakra-ui/react";

const colors = {
  brand: {
    50: "#e8eeff",
    100: "#c1cdf4",
    200: "#99abe7",
    300: "#718ada",
    400: "#4a68cf",
    500: "#304fb5",
    600: "#253d8e",
    700: "#192c67",
    800: "#0c1a40",
    900: "#01091b",
  },
};

const appTheme = extendTheme({ colors });

export default appTheme;
