import { FC } from "react";
import { z } from "zod";

import { Button, Stack } from "@chakra-ui/react";

import Form from "../../common/form/form";
import FormTextField from "../../common/form/form-text-field";
import FormSelectField from "../../common/form/form-select-field";
import pingSchema from "./ping.schema";
import { ServerLocation } from "../../api/get-api-url";

type PingFormProps = {
  onSubmit: (values: z.infer<typeof pingSchema>) => void;
  disabled?: boolean;
};

const PingForm: FC<PingFormProps> = ({ onSubmit, disabled }) => {
  return (
    <Form
      schema={pingSchema}
      initialValues={{ host: "" }}
      onSubmit={onSubmit}
      render={(form) => (
        <Stack align="flex-end" spacing={3}>
          <FormTextField
            name="host"
            isDisabled={disabled}
            variant="filled"
            label="Hostname or IP address"
            placeholder="Enter the address that you want to ping"
          />

          <FormSelectField
            name="location"
            isDisabled={disabled}
            variant="filled"
            label="Server location"
          >
            <option value={ServerLocation.LONDON}>London</option>
            <option value={ServerLocation.US_EAST}>US East</option>
            <option value={ServerLocation.US_WEST}>US West</option>
          </FormSelectField>

          <Button
            colorScheme="green"
            type="submit"
            px={6}
            isDisabled={disabled}
          >
            Start Pinging
          </Button>
        </Stack>
      )}
    />
  );
};

export default PingForm;
