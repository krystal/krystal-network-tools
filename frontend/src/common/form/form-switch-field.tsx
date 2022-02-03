import { forwardRef, ComponentPropsWithoutRef } from "react";

import { useField } from "react-final-form";
import {
  Select,
  FormControl,
  FormLabel,
  FormErrorMessage,
  HStack,
  Switch,
} from "@chakra-ui/react";

import { getFieldError } from "./form.helpers";

type FormSwitchFieldProps = ComponentPropsWithoutRef<typeof Select> & {
  name: string;
  label: string;
};

const FormSwitchField = forwardRef<HTMLInputElement, FormSwitchFieldProps>(
  ({ name, label, children, ...props }, ref) => {
    const { input, meta } = useField(name, {
      type: "checkbox",
    });

    const normalizedError = getFieldError(meta);

    return (
      <FormControl isInvalid={meta.touched && normalizedError}>
        <HStack>
          <Switch
            {...input}
            isChecked={input.checked}
            disabled={meta.submitting}
            {...props}
            ref={ref}
          />

          {label && (
            <FormLabel mb="0" htmlFor={input.id}>
              {label}
            </FormLabel>
          )}
        </HStack>
        <FormErrorMessage>{normalizedError}</FormErrorMessage>
      </FormControl>
    );
  }
);

export default FormSwitchField;
