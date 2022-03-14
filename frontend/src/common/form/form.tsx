import { Alert, AlertTitle, Box, Stack } from "@chakra-ui/react";
import { Form as FinalForm } from "react-final-form";
import { FORM_ERROR } from "final-form";
import { formValidator } from "./form.helpers";

import { FormComponent } from "./form.types";
import { useEffect, useState } from "react";

const Form: FormComponent = ({
  schema,
  initialValues: initialVals,
  onSubmit,
  render,
  ...props
}) => {
  // handleSubmit is a closure around the form submission function
  let handleSubmit: () => void;

  const [initialValues, setInitialValues] = useState(initialVals || {});

  // Check if the initial values can be added from the URL search params
  // We only want this to run on mount
  useEffect(() => {
    if (window.location.search) {
      const params = new URLSearchParams(window.location.search);

      setInitialValues(
        Object.entries(initialValues).reduce((values, [name, val]) => {
          let param: string | boolean | null = params.get(name);
          if (param === "false") param = false;
          if (param === "true") param = true;

          const value = param === null ? val : param;

          return { ...values, [name]: value };
        }, {})
      );
    }
  }, []); // eslint-disable-line

  // When the initial values are set check if they are valid
  // if they are then submit the form right away.
  useEffect(() => {
    if (initialValues !== initialVals) {
      try {
        schema.parse(initialValues);
        if (handleSubmit) handleSubmit();
      } catch (err) {
        // Do nothing.
      }
    }
  }, [initialValues]); // eslint-disable-line

  return (
    <FinalForm
      keepDirtyOnReinitialize
      initialValues={initialValues}
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
      render={(form) => {
        handleSubmit = form.handleSubmit;

        return (
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
        );
      }}
    />
  );
};

export default Form;
