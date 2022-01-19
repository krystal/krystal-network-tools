import { Icon, IconProps } from "@chakra-ui/react";
import { FC } from "react";

const LogoIcon: FC<IconProps> = (props) => (
  <Icon viewBox="0 0 135 133" {...props}>
    <g
      id="Page-1"
      stroke="none"
      stroke-width="1"
      fill="none"
      fill-rule="evenodd"
    >
      <g id="Artboard" fill="currentColor" fill-rule="nonzero">
        <g id="Group">
          <polygon id="Path" points="62 133 62 56.3333333 1 41"></polygon>
          <polygon
            id="Path"
            points="31.1907583 0 0 31.2 67.5 48 135 31.2 103.809242 0"
          ></polygon>
          <polygon id="Path" points="71 133 133 41 71 56.3333333"></polygon>
        </g>
      </g>
    </g>
  </Icon>
);

export default LogoIcon;
