import { FC } from "react";
import { z } from "zod";

import { Button, Stack } from "@chakra-ui/react";

import Form from "../../common/form/form";
import FormTextField from "../../common/form/form-text-field";
import whoisSchema from "./whois.schema";

type WhoisFormProps = {
  onSubmit: (values: z.infer<typeof whoisSchema>) => void;
  disabled?: boolean;
};

const WhoisForm: FC<WhoisFormProps> = ({ onSubmit, disabled }) => {
  return (
    <Form
      schema={whoisSchema}
      initialValues={{ host: "" }}
      onSubmit={onSubmit}
      render={(form) => (
        <Stack align="flex-end" spacing={3}>
          <FormTextField
            name="host"
            isDisabled={disabled}
            variant="filled"
            label="Host name or IP address"
            placeholder="Enter the address that you want to query"
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

export default WhoisForm;
