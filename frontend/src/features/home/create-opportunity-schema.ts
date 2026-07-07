import { z } from "zod"

export const createOpportunitySchema = z.object({
  company: z.string().min(1, "Company is required"),
  role: z.string().min(1, "Role is required"),
  description: z.string().min(1, "Description is required"),
})

export type CreateOpportunityFormValues = z.infer<typeof createOpportunitySchema>