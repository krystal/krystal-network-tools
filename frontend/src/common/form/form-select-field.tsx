import { forwardRef, ComponentPropsWithoutRef } from "react";

import { useField } from "react-final-form";
import {
  InputGroup,
  Select,
  FormControl,
  FormLabel,
  FormErrorMessage,
} from "@chakra-ui/react";

import { getFieldError } from "./form.helpers";

type FormSelectFieldProps = ComponentPropsWithoutRef<typeof Select> & {
  name: string;
  label: string;
};

const FormSelectField = forwardRef<HTMLSelectElement, FormSelectFieldProps>(
  ({ name, label, children, ...props }, ref) => {
    const { input, meta } = useField(name);

    const normalizedError = getFieldError(meta);

    return (
      <FormControl isInvalid={meta.touched && normalizedError}>
        {label && <FormLabel>{label}</FormLabel>}
        <InputGroup>
          <Select
            {...input}
            value={input.value}
            disabled={meta.submitting}
            {...props}
            ref={ref}
          >
            {children}
          </Select>
        </InputGroup>
        <FormErrorMessage>{normalizedError}</FormErrorMessage>
      </FormControl>
    );
  }
);

export default FormSelectField;
