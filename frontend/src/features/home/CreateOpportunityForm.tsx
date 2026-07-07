import { useForm, type Resolver } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2 } from "lucide-react"

import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Button } from "@/components/ui/button"
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from "@/components/ui/form"
import { S } from "@/lib/strings"
import { useCreateOpportunity } from "@/api/opportunities"
import {
  createOpportunitySchema,
  type CreateOpportunityFormValues,
} from "@/features/home/create-opportunity-schema"

interface CreateOpportunityFormProps {
  onCancel: () => void
  onCreated: () => void
}

// @orkai:ref(id=2cf97580-172f-410d-81b4-edb7e177a7b3)
// @orkai:ref(id=359cbba1-0559-4db9-af44-3533e0ab918c)
// @orkai:ref(id=5759c69e-7fc5-4e0c-8e9e-554fd4388492)
// @orkai:decision Inline Card-based form (no separate route) — App.tsx switches on gate not URL, so the create flow lives inside the Home page surface
function CreateOpportunityForm({ onCancel, onCreated }: CreateOpportunityFormProps) {
  const createOpportunity = useCreateOpportunity()

  const form = useForm<CreateOpportunityFormValues>({
    resolver: zodResolver(createOpportunitySchema) as unknown as Resolver<CreateOpportunityFormValues>,
    defaultValues: {
      company: "",
      role: "",
      description: "",
    },
  })

  function onSubmit(values: CreateOpportunityFormValues) {
    createOpportunity.mutate(values, {
      onSuccess: (res) => {
        if (!res.error) {
          onCreated()
        }
      },
    })
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="company"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{S.home.createCompanyLabel}</FormLabel>
              <FormControl>
                <Input placeholder={S.home.createCompanyPlaceholder} {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="role"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{S.home.createRoleLabel}</FormLabel>
              <FormControl>
                <Input placeholder={S.home.createRolePlaceholder} {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="description"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{S.home.createDescriptionLabel}</FormLabel>
              <FormControl>
                <Textarea
                  placeholder={S.home.createDescriptionPlaceholder}
                  rows={6}
                  {...field}
                  onPaste={(e) => {
                    const pasted = e.clipboardData.getData("text/plain")
                    if (!pasted) return

                    e.preventDefault()

                    const cleaned = pasted
                      .replace(/\r\n/g, "\n")
                      .replace(/\r/g, "\n")
                      .replace(/\n{3,}/g, "\n\n")
                      .trim()

                    const target = e.currentTarget
                    const start = target.selectionStart
                    const end = target.selectionEnd
                    const currentValue = field.value || ""
                    const newValue =
                      currentValue.slice(0, start) +
                      cleaned +
                      currentValue.slice(end)

                    field.onChange(newValue)

                    requestAnimationFrame(() => {
                      target.selectionStart = target.selectionEnd =
                        start + cleaned.length
                    })
                  }}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={onCancel}>
            {S.home.createCancel}
          </Button>
          <Button type="submit" disabled={createOpportunity.isPending}>
            {createOpportunity.isPending ? (
              <>
                <Loader2 className="mr-2 size-4 animate-spin" />
                {S.home.createSubmit}
              </>
            ) : (
              S.home.createSubmit
            )}
          </Button>
        </div>
      </form>
    </Form>
  )
}

export { CreateOpportunityForm }
export type { CreateOpportunityFormProps }