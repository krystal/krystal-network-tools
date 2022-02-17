import { FC } from "react";
import { z } from "zod";

import { Button, Stack } from "@chakra-ui/react";

import Form from "../../common/form/form";
import FormTextField from "../../common/form/form-text-field";
import FormSelectField from "../../common/form/form-select-field";
import pingSchema from "./ping.schema";
import { locations } from "../../api/server-locations";

type PingFormProps = {
  onSubmit: (values: z.infer<typeof pingSchema>) => void;
  disabled?: boolean;
};

const PingForm: FC<PingFormProps> = ({ onSubmit, disabled }) => {
  return (
    <Form
      schema={pingSchema}
      initialValues={{ host: "", location: locations[0].id }}
      onSubmit={onSubmit}
      render={(form) => (
        <Stack align="flex-end" spacing={3}>
          <FormTextField
            name="host"
            isDisabled={disabled}
            label="Hostname or IP address"
            placeholder="Enter the address that you want to ping"
          />

          <FormSelectField
            name="location"
            isDisabled={disabled}
            label="Server location"
          >
            {
              locations.map((location, index) => <option key={index} value={location.id}>
                {location.name}
              </option>)
            }
          </FormSelectField>

          <Button
            colorScheme="green"
            type="submit"
            px={6}
            isDisabled={disabled}
          >
            Start Ping
          </Button>
        </Stack>
      )}
    />
  );
};

export default PingForm;
