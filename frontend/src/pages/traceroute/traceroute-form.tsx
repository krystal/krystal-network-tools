import { FC } from "react";
import { z } from "zod";

import { Button, Stack } from "@chakra-ui/react";

import Form from "../../common/form/form";
import FormTextField from "../../common/form/form-text-field";
import tracerouteSchema from "./traceroute.schema";

type TracerouteFormProps = {
  onSubmit: (values: z.infer<typeof tracerouteSchema>) => void;
  disabled?: boolean;
};

const TracerouteForm: FC<TracerouteFormProps> = ({ onSubmit, disabled }) => {
  return (
    <Form
      schema={tracerouteSchema}
      initialValues={{ host: "" }}
      onSubmit={onSubmit}
      render={(form) => (
        <Stack align="flex-end" spacing={3}>
          <FormTextField
            name="host"
            isDisabled={disabled}
            label="Hostname or IP address"
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

export default TracerouteForm;
