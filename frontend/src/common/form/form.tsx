import { Alert, AlertTitle, Box, Stack } from "@chakra-ui/react";
import { Form as FinalForm } from "react-final-form";
import { FORM_ERROR } from "final-form";
import { formValidator } from "./form.helpers";

import { FormComponent } from "./form.types";

const Form: FormComponent = ({
  schema,
  initialValues = {},
  onSubmit,
  render,
  ...props
}) => {
  return (
    <FinalForm
      keepDirtyOnReinitialize
      // Check if the initial values can be added from the URL search params
      initialValues={Object.keys(initialValues).reduce((values, name) => {
        const params = new URLSearchParams(window.location.search);
        let param: string | boolean | null = params.get(name);
        if (param === "false") param = false;
        if (param === "true") param = true;
        const value = param || (initialValues as any)[name];
        return { ...values, [name]: value };
      }, {})}
      validate={formValidator(schema)}
      onSubmit={async (...args) => {
        try {
          await onSubmit(...args);
          const values = new URLSearchParams(args[0]);
          // On a successful commit, add the form values to the search params
          window.history.pushState(
            {},
            window.location.pathname,
            window.location.origin +
              window.location.pathname +
              "?" +
              values.toString()
          );
        } catch (err) {
          console.log({ err });
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
