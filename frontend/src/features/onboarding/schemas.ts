import { z } from "zod"

export const workExperienceSchema = z.object({
  jobTitle: z.string().min(1, "Job title is required"),
  company: z.string().min(1, "Company is required"),
  location: z.string().min(1, "Location is required"),
  startDate: z.string().min(1, "Start date is required"),
  endDate: z.string().min(1, "End date is required"),
  description: z.string().optional().default(""),
})

export const educationSchema = z.object({
  degree: z.string().min(1, "Degree is required"),
  institution: z.string().min(1, "Institution is required"),
  location: z.string().min(1, "Location is required"),
  startDate: z.string().min(1, "Start date is required"),
  endDate: z.string().min(1, "End date is required"),
  gpa: z.string().optional(),
  description: z.string().optional().default(""),
})

export const skillCategorySchema = z.object({
  name: z.string().min(1, "Category name is required"),
  skills: z.array(z.string().min(1, "Skill is required")).min(1, "At least one skill is required"),
})

export const projectSchema = z.object({
  name: z.string().min(1, "Project name is required"),
  role: z.string().min(1, "Role is required"),
  description: z.string().min(1, "Description is required"),
  technologies: z.array(z.string().min(1, "Technology is required")).min(1, "At least one technology is required"),
  url: z.string().url("Must be a valid URL").optional().or(z.literal("")),
})

export const certificationSchema = z.object({
  name: z.string().min(1, "Certification name is required"),
  issuingOrg: z.string().min(1, "Issuing organization is required"),
  dateObtained: z.string().min(1, "Date obtained is required"),
  expiryDate: z.string().optional(),
  credentialUrl: z.string().url("Must be a valid URL").optional().or(z.literal("")),
})

export const languageSchema = z.object({
  name: z.string().min(1, "Language name is required"),
  proficiency: z.string().min(1, "Proficiency is required"),
})

export const profileSchema = z.object({
  fullName: z.string().min(1, "Full name is required"),
  email: z.string().email("Invalid email address").min(1, "Email is required"),
  phone: z.string().optional().default(""),
  location: z.string().optional().default(""),
  linkedinUrl: z.string().url("Must be a valid URL").optional().or(z.literal("")),
  websiteUrl: z.string().url("Must be a valid URL").optional().or(z.literal("")),
  githubUrl: z.string().url("Must be a valid URL").optional().or(z.literal("")),
  professionalSummary: z.string().optional().default(""),
  workExperience: z.array(workExperienceSchema).optional().default([]),
  education: z.array(educationSchema).optional().default([]),
  skills: z.array(skillCategorySchema).optional().default([]),
  projects: z.array(projectSchema).optional().default([]),
  certifications: z.array(certificationSchema).optional().default([]),
  languages: z.array(languageSchema).optional().default([]),
})

export const llmConfigSchema = z.object({
  provider: z.enum(["ollama", "openai", "anthropic"]),
  apiKey: z.string().optional().default(""),
  model: z.string().optional().default(""),
})

export type WorkExperienceFormValues = z.infer<typeof workExperienceSchema>
export type EducationFormValues = z.infer<typeof educationSchema>
export type SkillCategoryFormValues = z.infer<typeof skillCategorySchema>
export type ProjectFormValues = z.infer<typeof projectSchema>
export type CertificationFormValues = z.infer<typeof certificationSchema>
export type LanguageFormValues = z.infer<typeof languageSchema>
export type ProfileFormValues = z.infer<typeof profileSchema>
export type LLMConfigFormValues = z.infer<typeof llmConfigSchema>
