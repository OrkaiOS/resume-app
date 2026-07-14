import * as React from "react"
import {
  Controller,
  type ControllerProps,
  type FieldPath,
  type FieldValues,
  FormProvider,
  type UseFormReturn,
} from "react-hook-form"

import { cn } from "@/lib/utils"
import { Label } from "@/components/ui/label"
import {
  FormFieldContext,
  FormItemContext,
  useFormField,
} from "./use-form-field"

function Form<TFieldValues extends FieldValues>({
  children,
  ...props
}: {
  children: React.ReactNode
} & UseFormReturn<TFieldValues>) {
  return <FormProvider {...props}>{children}</FormProvider>
}

function FormField<
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({ ...props }: ControllerProps<TFieldValues, TName>) {
  return (
    <FormFieldContext.Provider value={{ name: props.name }}>
      <Controller {...props} />
    </FormFieldContext.Provider>
  )
}

interface FormItemProps extends React.ComponentProps<"div"> {}

function FormItem({ className, ...props }: FormItemProps) {
  const id = React.useId()

  return (
    <FormItemContext.Provider value={{ id }}>
      <div
        data-slot="form-item"
        className={cn("grid gap-1.5", className)}
        {...props}
      />
    </FormItemContext.Provider>
  )
}

interface FormLabelProps extends React.ComponentProps<typeof Label> {}

function FormLabel({ className, ...props }: FormLabelProps) {
  const { error, formItemId } = useFormField()

  return (
    <Label
      data-slot="form-label"
      className={cn("data-[error=true]:text-destructive", className)}
      htmlFor={formItemId}
      data-error={!!error}
      {...props}
    />
  )
}

interface FormControlProps extends React.ComponentProps<"div"> {}

function FormControl({ ...props }: FormControlProps) {
  const { error, formItemId, formDescriptionId, formMessageId } =
    useFormField()

  return (
    <div
      data-slot="form-control"
      id={formItemId}
      aria-describedby={
        !error
          ? `${formDescriptionId}`
          : `${formDescriptionId} ${formMessageId}`
      }
      aria-invalid={!!error}
      {...props}
    />
  )
}

interface FormDescriptionProps extends React.ComponentProps<"p"> {}

function FormDescription({ className, ...props }: FormDescriptionProps) {
  const { formDescriptionId } = useFormField()

  return (
    <p
      data-slot="form-description"
      id={formDescriptionId}
      className={cn("text-[0.8rem] text-muted-foreground", className)}
      {...props}
    />
  )
}

interface FormMessageProps extends React.ComponentProps<"p"> {}

function FormMessage({ className, children, ...props }: FormMessageProps) {
  const { error, formMessageId } = useFormField()
  const body = error ? String((error as { message?: string })?.message ?? "") : children

  if (!body) {
    return null
  }

  return (
    <p
      data-slot="form-message"
      id={formMessageId}
      className={cn("text-[0.8rem] font-medium text-destructive", className)}
      {...props}
    >
      {body}
    </p>
  )
}

export {
  Form,
  FormProvider,
  FormItem,
  FormLabel,
  FormControl,
  FormDescription,
  FormMessage,
  FormField,
}
