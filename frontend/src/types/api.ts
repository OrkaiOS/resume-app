export interface ApiEnvelope<T> {
  data: T
  error: null
}

export interface ApiError {
  data: null
  error: {
    code: string
    message: string
    details: Record<string, unknown>
  }
}

export type ApiResponse<T> = ApiEnvelope<T> | ApiError

export interface PaginatedResponse<T> {
  items: T[]
  nextCursor: string
}

export interface OrkaiHealth {
  connected: boolean
}

export interface WorkExperience {
  jobTitle: string
  company: string
  location: string
  startDate: string
  endDate: string
  description: string
}

export interface Education {
  degree: string
  institution: string
  location: string
  startDate: string
  endDate: string
  gpa?: string
  description: string
}

export interface SkillCategory {
  name: string
  skills: string[]
}

export interface Project {
  name: string
  role: string
  description: string
  technologies: string[]
  url: string
}

export interface Certification {
  name: string
  issuingOrg: string
  dateObtained: string
  expiryDate?: string
  credentialUrl?: string
}

export interface Language {
  name: string
  proficiency: string
}

export interface ProfileResponse {
  id: string
  fullName: string
  email: string
  phone: string
  location: string
  linkedinUrl: string
  websiteUrl: string
  githubUrl: string
  professionalSummary: string
  workExperience: WorkExperience[]
  education: Education[]
  skills: SkillCategory[]
  projects: Project[]
  certifications: Certification[]
  languages: Language[]
  createdAt: string
  updatedAt: string
}

export interface ProfileUpsertRequest {
  fullName: string
  email: string
  phone: string
  location: string
  linkedinUrl: string
  websiteUrl: string
  githubUrl: string
  professionalSummary: string
  workExperience: WorkExperience[]
  education: Education[]
  skills: SkillCategory[]
  projects: Project[]
  certifications: Certification[]
  languages: Language[]
}

export interface OpportunityResponse {
  id: string
  company: string
  role: string
  description: string
  status: string
  createdAt: string
  updatedAt: string
}

export interface OpportunityCreateRequest {
  company: string
  role: string
  description: string
}

export interface OpportunityUpdateRequest {
  company: string
  role: string
  description: string
  status: string
}

export interface OpportunityArchiveRequest {
  archived: boolean
}

export interface ResumeResponse {
  id: string
  opportunityId: string
  status: string
  markdownContent: string
  pdfPath: string
  createdAt: string
  updatedAt: string
}

export interface CoverLetterResponse {
  id: string
  opportunityId: string
  status: string
  markdownContent: string
  pdfPath: string
  createdAt: string
  updatedAt: string
}

export interface ChatSendRequest {
  opportunityId: string
  message: string
}

export type ChatEvent =
  | { type: "token"; content: string }
  | { type: "tool_call"; tool: string; input: unknown }
  | { type: "tool_result"; tool: string; output: unknown }
  | { type: "draft"; documentType: "resume" | "cover_letter"; markdown: string }
  | { type: "pdf_ready"; documentType: "resume" | "cover_letter"; url: string }
  | { type: "done" }
  | { type: "error"; code: string; message: string }

export interface ArtifactResponse {
  id: string
  name: string
  type: "python" | "bash"
  description: string
  content: string
  createdAt: string
  lastUsedAt: string
  usageCount: number
}

export interface OrkaiSearchResult {
  id: string
  name: string
  type: string
  snippet: string
}

export interface LLMConfigRequest {
  provider: "ollama" | "openai" | "anthropic"
  apiKey?: string
  model?: string
}

export interface LLMConfigResponse {
  validated: boolean
}

export interface OrkaiSetupStep {
  name: string
  status: "pending" | "in_progress" | "success" | "failed"
  error?: string
}

export interface OrkaiSetupStatus {
  setupId: string
  steps: OrkaiSetupStep[]
  completed: boolean
}
