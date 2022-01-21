import { Form as FinalForm } from "react-final-form";

import type {
  FormProps as FinalFormProps,
  FormRenderProps,
} from "react-final-form";
import type { TypeOf, z, ZodType } from "zod";
import type { PropsWithoutRef } from "react";

const validator =
  <T extends z.ZodType<any, any>>(schema: T) =>
  (values: any) => {
    try {
      schema.parse(values);
      return {};
    } catch (err) {
      return (err as z.ZodError).formErrors.fieldErrors;
    }
  };

type FormProps<S extends ZodType<any>> = PropsWithoutRef<
  Omit<JSX.IntrinsicElements["form"], "onSubmit">
> & {
  schema: S;
  onSubmit: FinalFormProps<z.infer<S>>["onSubmit"];
  initialValues?: FinalFormProps<z.infer<S>>["initialValues"];
  render: (props: FormRenderProps<TypeOf<S>>) => JSX.Element;
};

export type FormComponent<P = {}> = <S extends ZodType<any>>(
  props: FormProps<S> & P
) => JSX.Element;

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
      validate={validator(schema)}
      onSubmit={onSubmit}
      render={(form) => (
        <form onSubmit={form.handleSubmit} {...props}>
          {render(form)}
        </form>
      )}
    />
  );
};

export default Form;
