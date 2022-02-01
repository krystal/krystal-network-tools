import { FC } from "react";
import { z } from "zod";

import { Button, Stack } from "@chakra-ui/react";

import Form from "../../common/form/form";
import FormTextField from "../../common/form/form-text-field";
import bgpRouteSchema from "./bgp-route.schema";

type BgpRouteFormProps = {
  onSubmit: (values: z.infer<typeof bgpRouteSchema>) => void;
  disabled?: boolean;
};

const BgpRouteForm: FC<BgpRouteFormProps> = ({ onSubmit, disabled }) => {
  return (
    <Form
      schema={bgpRouteSchema}
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

export default BgpRouteForm;
