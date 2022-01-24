import { Form as FinalForm } from "react-final-form";
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
