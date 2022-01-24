import { FieldMetaState } from "react-final-form";
import type { z } from "zod";

export const formValidator =
  <T extends z.ZodType<any, any>>(schema: T) =>
  (values: any) => {
    try {
      schema.parse(values);
      return {};
    } catch (err) {
      return (err as z.ZodError).formErrors.fieldErrors;
    }
  };

export const getFieldError = (meta: FieldMetaState<any>) => {
  return Array.isArray(meta.error)
    ? meta.error.join(", ")
    : meta.error || meta.submitError;
};
