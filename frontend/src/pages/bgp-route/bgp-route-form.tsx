import { FC } from "react";
import { z } from "zod";

import { Button, Stack } from "@chakra-ui/react";

import Form from "../../common/form/form";
import FormTextField from "../../common/form/form-text-field";
import bgpRouteSchema from "./bgp-route.schema";
import FormSelectField from "../../common/form/form-select-field";
import { locations } from "../../api/server-locations";

type BgpRouteFormProps = {
  onSubmit: (values: z.infer<typeof bgpRouteSchema>) => void;
  disabled?: boolean;
};

const BgpRouteForm: FC<BgpRouteFormProps> = ({ onSubmit, disabled }) => {
  return (
    <Form
      schema={bgpRouteSchema}
      initialValues={{ ip: "", location: locations[0].id }}
      onSubmit={onSubmit}
      render={(form) => (
        <Stack align="flex-end" spacing={3}>
          <FormTextField
            name="ip"
            isDisabled={disabled}
            label="IP address"
            placeholder="Enter the address that you want to lookup"
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

export default BgpRouteForm;
