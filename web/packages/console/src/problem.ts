import type { FieldErrorItem, FormRefusal } from "@cdr-lab/ui/form";
import type { Problem } from "@cdr-lab/ui/states";

// The API document of each service defines its own problem type for field errors.
export type FieldErrorsProblem = {
  type: string;
  title: string;
  status: number;
  detail: string;
  errors: FieldErrorItem[];
};

export function hasFieldErrors(
  problem: Problem | FieldErrorsProblem,
  fieldErrorsType: string,
): problem is FieldErrorsProblem {
  return "type" in problem && problem.type === fieldErrorsType;
}

export class ProblemError extends Error {
  readonly problem: Problem;
  readonly status: number;
  readonly fieldErrors: readonly FieldErrorItem[];

  constructor(problem: Problem | FieldErrorsProblem, status: number, fieldErrorsType: string) {
    super(problem.title);
    this.problem = problem;
    this.status = status;
    this.fieldErrors = hasFieldErrors(problem, fieldErrorsType) ? problem.errors : [];
  }
}

// An error that is not a problem goes to the form, which shows a failure with no problem details.
export function refusalOf(error: unknown): FormRefusal {
  if (!(error instanceof ProblemError)) throw error;
  return { problem: error.problem, fieldErrors: error.fieldErrors };
}

// The service is the name in a sentence, for example "the Register".
export function problemOf(error: unknown, service: string): Problem {
  return error instanceof ProblemError
    ? error.problem
    : {
        title: `The console cannot get data from ${service}`,
        detail: `Make sure that ${service} runs, and then try again.`,
      };
}
