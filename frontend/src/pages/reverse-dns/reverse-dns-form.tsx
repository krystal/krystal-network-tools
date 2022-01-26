import { FC } from "react";
import { z } from "zod";

import { Button, Stack } from "@chakra-ui/react";

import Form from "../../common/form/form";
import FormTextField from "../../common/form/form-text-field";
import reverseDnsSchema from "./reverse-dns.schema";

type ReverseDnsFormProps = {
  onSubmit: (values: z.infer<typeof reverseDnsSchema>) => void;
  disabled?: boolean;
};

const ReverseDnsForm: FC<ReverseDnsFormProps> = ({ onSubmit, disabled }) => {
  return (
    <Form
      schema={reverseDnsSchema}
      initialValues={{ ip: "" }}
      onSubmit={onSubmit}
      render={(form) => (
        <Stack align="flex-end" spacing={3}>
          <FormTextField
            name="ip"
            isDisabled={disabled}
            label="IP address"
            placeholder="Enter the address that you want to lookup"
          />

          <Button
            colorScheme="green"
            type="submit"
            px={6}
            isLoading={form.submitting}
            isDisabled={disabled}
          >
            Submit
          </Button>
        </Stack>
      )}
    />
  );
};

export default ReverseDnsForm;
