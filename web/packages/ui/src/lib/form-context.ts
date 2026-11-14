import { createContext, useContext } from "react";

export type FormContextValue = {
  idOf: (name: string) => string;
  errorOf: (name: string) => string | undefined;
};

export const FormContext = createContext<FormContextValue | undefined>(undefined);

// In a form, the name of a field is the JSON Pointer path of its value in the request body, for example
// "endpointDetail/publicBaseUri". The form gives the field its id and its errors from the server by that name.
export function useFormBinding(
  name: string | undefined,
  error: string | undefined,
  id: string | undefined,
): { error: string | undefined; id: string | undefined } {
  const form = useContext(FormContext);
  if (!form || !name) return { error, id };
  return { error: error ?? form.errorOf(name), id: id ?? form.idOf(name) };
}
