import { FC } from "react";
import { z } from "zod";

import { Button, Stack } from "@chakra-ui/react";

import Form from "../../common/form/form";
import FormTextField from "../../common/form/form-text-field";
import FormSelectField from "../../common/form/form-select-field";
import dnsSchema, { DnsType } from "./dns.schema";
import FormSwitchField from "../../common/form/form-switch-field";

type DnsFormProps = {
  onSubmit: (values: z.infer<typeof dnsSchema>) => void;
  disabled?: boolean;
};

const DnsForm: FC<DnsFormProps> = ({ onSubmit, disabled }) => {
  return (
    <Form
      schema={dnsSchema}
      initialValues={{ host: "", type: DnsType.ANY, trace: false, cache: false }}
      onSubmit={onSubmit}
      render={(form) => (
        <Stack align="flex-end" spacing={3}>
          <FormTextField
            name="host"
            isDisabled={disabled}
            label="Hostname or IP address"
            placeholder="Enter the address that you want to query"
          />

          <FormSelectField
            name="type"
            isDisabled={disabled}
            label="Record type"
          >
            <option value={DnsType.ANY}>ANY</option>
            <option value={DnsType.A}>A</option>
            <option value={DnsType.AAAA}>AAAA</option>
            <option value={DnsType.MX}>MX</option>
            <option value={DnsType.NS}>NS</option>
            <option value={DnsType.PTR}>PTR</option>
            <option value={DnsType.SOA}>SOA</option>
            <option value={DnsType.SRV}>SRV</option>
            <option value={DnsType.TXT}>TXT</option>
          </FormSelectField>

          <FormSwitchField name="trace" label="Full trace?" />
          <FormSwitchField name="cache" label="Use cache?" />

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

export default DnsForm;
