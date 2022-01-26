import { Alert, AlertTitle, Box, Stack } from "@chakra-ui/react";
import { Form as FinalForm } from "react-final-form";
import { FORM_ERROR } from "final-form";
import { formValidator } from "./form.helpers";

import { FormComponent } from "./form.types";

const Form: FormComponent = ({
  schema,
  initialValues,
  onSubmit,
  render,
  ...props
}) => {
  return (
    <FinalForm
      initialValues={initialValues}
      validate={formValidator(schema)}
      onSubmit={async (...args) => {
        try {
          await onSubmit(...args);
        } catch (err) {
          return {
            [FORM_ERROR]:
              err instanceof Error
                ? err.message
                : "There was a problem submitting the form. Please check and try again.",
          };
        }
      }}
      render={(form) => (
        <form onSubmit={form.handleSubmit} {...props}>
          <Stack spacing={6}>
            {form.submitError && (
              <Alert status="error" variant="solid" borderRadius="sm">
                <Box flex="1">
                  <AlertTitle>
                    There was a problem submitting the form
                  </AlertTitle>
                  {form.submitError}
                </Box>
              </Alert>
            )}

            {render(form)}
          </Stack>
        </form>
      )}
    />
  );
};

export default Form;
