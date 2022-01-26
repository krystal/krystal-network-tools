import { forwardRef, ComponentPropsWithoutRef } from "react";

import { useField } from "react-final-form";
import { Input, InputGroup } from "@chakra-ui/input";
import {
  FormControl,
  FormLabel,
  FormErrorMessage,
} from "@chakra-ui/form-control";
import { getFieldError } from "./form.helpers";

type FormTextFieldProps = ComponentPropsWithoutRef<typeof Input> & {
  name: string;
  label: string;
  type?: "text" | "password" | "email" | "number";
};

const FormTextField = forwardRef<HTMLInputElement, FormTextFieldProps>(
  ({ name, label, ...props }, ref) => {
    const { input, meta } = useField(name, {
      parse:
        props.type === "number"
          ? (v) => (v === "" ? null : Number(v))
          : (v) => (v === "" ? null : v),
    });

    const normalizedError = getFieldError(meta);

    return (
      <FormControl isInvalid={meta.touched && normalizedError}>
        {label && <FormLabel>{label}</FormLabel>}
        <InputGroup>
          <Input
            {...input}
            disabled={meta.submitting}
            variant="filled"
            {...props}
            ref={ref}
          />
        </InputGroup>
        <FormErrorMessage>{normalizedError}</FormErrorMessage>
      </FormControl>
    );
  }
);

export default FormTextField;
