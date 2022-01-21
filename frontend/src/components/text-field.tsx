import { forwardRef, ComponentPropsWithoutRef } from "react";

import { useField } from "react-final-form";
import { Input, InputGroup } from "@chakra-ui/input";
import {
  FormControl,
  FormLabel,
  FormErrorMessage,
} from "@chakra-ui/form-control";

type TextFieldProps = ComponentPropsWithoutRef<typeof Input> & {
  name: string;
  label: string;
  type?: "text" | "password" | "email" | "number";
};

const TextField = forwardRef<HTMLInputElement, TextFieldProps>(
  ({ name, label, ...props }, ref) => {
    const { input, meta } = useField(name, {
      parse:
        props.type === "number"
          ? (v) => (v === "" ? null : Number(v))
          : (v) => (v === "" ? null : v),
    });

    const { touched, error, submitError, submitting } = meta;

    const normalizedError = Array.isArray(error)
      ? error.join(", ")
      : error || submitError;

    return (
      <FormControl isInvalid={touched && normalizedError}>
        {label && <FormLabel>{label}</FormLabel>}
        <InputGroup>
          <Input {...input} disabled={submitting} {...props} ref={ref} />
        </InputGroup>
        <FormErrorMessage>{normalizedError}</FormErrorMessage>
      </FormControl>
    );
  }
);

export default TextField;
