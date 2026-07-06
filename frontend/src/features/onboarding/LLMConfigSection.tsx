import { useForm, type Resolver } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2 } from "lucide-react"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from "@/components/ui/form"

import { llmConfigSchema, type LLMConfigFormValues } from "@/features/onboarding/schemas"
import { useSaveLLMConfig } from "@/api/onboarding"

interface LLMConfigSectionProps {
  onComplete: () => void
}

function LLMConfigSection({ onComplete }: LLMConfigSectionProps) {
  const saveLLMConfig = useSaveLLMConfig()

  const form = useForm<LLMConfigFormValues>({
    resolver: zodResolver(llmConfigSchema) as unknown as Resolver<LLMConfigFormValues>,
    defaultValues: {
      provider: "ollama",
      apiKey: "",
      model: "",
    },
  })

  const provider = form.watch("provider")

  function onSubmit(values: LLMConfigFormValues) {
    saveLLMConfig.mutate(values, {
      onSuccess: () => onComplete(),
    })
  }

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 pt-2">
      <FormField
        control={form.control}
        name="provider"
        render={({ field }) => (
          <FormItem>
            <FormLabel>LLM Provider</FormLabel>
            <Select value={field.value} onValueChange={field.onChange}>
              <FormControl>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Select provider" />
                </SelectTrigger>
              </FormControl>
              <SelectContent>
                <SelectItem value="ollama">Ollama (local)</SelectItem>
                <SelectItem value="openai">OpenAI</SelectItem>
                <SelectItem value="anthropic">Anthropic</SelectItem>
              </SelectContent>
            </Select>
            <FormMessage />
          </FormItem>
        )}
      />

      {provider !== "ollama" && (
        <FormField
          control={form.control}
          name="apiKey"
          render={({ field }) => (
            <FormItem>
              <FormLabel>API Key</FormLabel>
              <FormControl>
                <Input type="password" placeholder="Enter your API key" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      )}

      <FormField
        control={form.control}
        name="model"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Model (optional)</FormLabel>
            <FormControl>
              <Input placeholder="Default model name" {...field} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />

      <Button type="submit" disabled={saveLLMConfig.isPending} className="w-full">
        {saveLLMConfig.isPending ? (
          <>
            <Loader2 className="mr-2 size-4 animate-spin" />
            Saving...
          </>
        ) : (
          "Save & Continue"
        )}
      </Button>
    </form>
  )
}

export { LLMConfigSection }
export type { LLMConfigSectionProps }
