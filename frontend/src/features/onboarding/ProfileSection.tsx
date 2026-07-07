import { useState, useCallback, useRef } from "react"
import { useForm, useFieldArray, type Resolver } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2, Plus, Trash2, Upload } from "lucide-react"

import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Label } from "@/components/ui/label"
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from "@/components/ui/form"

import { profileSchema, type ProfileFormValues } from "@/features/onboarding/schemas"
import { useSaveProfile, useUploadProfile } from "@/api/onboarding"
import type { ProfileUpsertRequest } from "@/types/api"

function toProfileRequest(values: ProfileFormValues): ProfileUpsertRequest {
  return {
    ...values,
    linkedinUrl: values.linkedinUrl || "",
    websiteUrl: values.websiteUrl || "",
    githubUrl: values.githubUrl || "",
    projects: values.projects.map((p) => ({
      ...p,
      url: p.url || "",
    })),
  }
}

function sanitizeProfileResponse(data: Record<string, unknown>): ProfileFormValues {
  const sanitizeStr = (v: unknown) => (typeof v === "string" ? v : "") ?? ""
  const fixURL = (v: unknown) => {
    const s = sanitizeStr(v)
    if (s && !/^https?:\/\//i.test(s)) return "https://" + s
    return s
  }

  return {
    fullName: sanitizeStr(data.fullName),
    email: sanitizeStr(data.email),
    phone: sanitizeStr(data.phone),
    location: sanitizeStr(data.location),
    linkedinUrl: fixURL(data.linkedinUrl),
    websiteUrl: fixURL(data.websiteUrl),
    githubUrl: fixURL(data.githubUrl),
    professionalSummary: sanitizeStr(data.professionalSummary),
    workExperience: normalizeArray(data.workExperience).map(toWorkExperience),
    education: normalizeArray(data.education).map(toEducation),
    skills: normalizeArray(data.skills).map(toSkillCategory),
    projects: normalizeArray(data.projects).map(toProject),
    certifications: normalizeArray(data.certifications).map(toCertification),
    languages: normalizeArray(data.languages).map(toLanguage),
  }
}

function normalizeArray(v: unknown): Record<string, unknown>[] {
  return Array.isArray(v) ? (v as Record<string, unknown>[]) : []
}

function toWorkExperience(w: Record<string, unknown>) {
  return {
    jobTitle: str(w.jobTitle),
    company: str(w.company),
    location: str(w.location),
    startDate: str(w.startDate),
    endDate: str(w.endDate),
    description: str(w.description),
  }
}

function toEducation(e: Record<string, unknown>) {
  return {
    degree: str(e.degree),
    institution: str(e.institution),
    location: str(e.location),
    startDate: str(e.startDate),
    endDate: str(e.endDate),
    gpa: str(e.gpa),
    description: str(e.description),
  }
}

function toSkillCategory(s: Record<string, unknown>) {
  return {
    name: str(s.name),
    skills: Array.isArray(s.skills) ? (s.skills as string[]).filter((x) => typeof x === "string") : [],
  }
}

function toProject(p: Record<string, unknown>) {
  const rawURL = str(p.url)
  return {
    name: str(p.name),
    role: str(p.role),
    description: str(p.description),
    technologies: Array.isArray(p.technologies)
      ? (p.technologies as string[]).filter((x) => typeof x === "string")
      : [],
    url: rawURL && !/^https?:\/\//i.test(rawURL) ? "https://" + rawURL : rawURL,
  }
}

function toCertification(c: Record<string, unknown>) {
  return {
    name: str(c.name),
    issuingOrg: str(c.issuingOrg),
    dateObtained: str(c.dateObtained),
    expiryDate: str(c.expiryDate),
    credentialUrl: str(c.credentialUrl),
  }
}

function toLanguage(l: Record<string, unknown>) {
  return {
    name: str(l.name),
    proficiency: str(l.proficiency),
  }
}

function str(v: unknown): string {
  return typeof v === "string" ? v : ""
}

const PROFILE_STORAGE_KEY = "onboarding-profile-form"

interface ProfileSectionProps {
  onComplete: () => void
}

function loadPersistedProfile(): ProfileFormValues | undefined {
  try {
    const stored = localStorage.getItem(PROFILE_STORAGE_KEY)
    if (stored) return JSON.parse(stored) as ProfileFormValues
  } catch {
    return undefined
  }
}

function persistProfile(values: ProfileFormValues) {
  try {
    localStorage.setItem(PROFILE_STORAGE_KEY, JSON.stringify(values))
  } catch {
    // ignore storage errors
  }
}

function ProfileSection({ onComplete }: ProfileSectionProps) {
  const [activeTab, setActiveTab] = useState<"manual" | "upload">("manual")
  const [dragActive, setDragActive] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const saveProfile = useSaveProfile()
  const uploadProfile = useUploadProfile()

  const persisted = loadPersistedProfile()

  const form = useForm<ProfileFormValues>({
    mode: "onBlur",
    resolver: zodResolver(profileSchema) as unknown as Resolver<ProfileFormValues>,
    defaultValues: persisted ?? {
      fullName: "",
      email: "",
      phone: "",
      location: "",
      linkedinUrl: "",
      websiteUrl: "",
      githubUrl: "",
      professionalSummary: "",
      workExperience: [],
      education: [],
      skills: [],
      projects: [],
      certifications: [],
      languages: [],
    },
  })

  const workExperienceArray = useFieldArray({ control: form.control, name: "workExperience" })
  const educationArray = useFieldArray({ control: form.control, name: "education" })
  const skillsArray = useFieldArray({ control: form.control, name: "skills" })
  const { update: updateSkills } = skillsArray
  const projectsArray = useFieldArray({ control: form.control, name: "projects" })
  const { update: updateProjects } = projectsArray
  const certificationsArray = useFieldArray({ control: form.control, name: "certifications" })
  const languagesArray = useFieldArray({ control: form.control, name: "languages" })

  const handleChange = useCallback(() => {
    const values = form.getValues()
    persistProfile(values)
  }, [form])

  function onSubmit(values: ProfileFormValues) {
    saveProfile.mutate(toProfileRequest(values), {
      onSuccess: () => {
        localStorage.removeItem(PROFILE_STORAGE_KEY)
        onComplete()
      },
      onError: () => {
        // toast handled by hook
      },
    })
  }

  function handleDrag(e: React.DragEvent) {
    e.preventDefault()
    e.stopPropagation()
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true)
    } else if (e.type === "dragleave") {
      setDragActive(false)
    }
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault()
    e.stopPropagation()
    setDragActive(false)
    const file = e.dataTransfer.files?.[0]
    if (file) handleFile(file)
  }

  function handleFileSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (file) handleFile(file)
  }

  function handleFile(file: File) {
    uploadProfile.mutate(file, {
      onSuccess: (res) => {
        if (res.data && !res.error) {
          form.reset(sanitizeProfileResponse(res.data as unknown as Record<string, unknown>))
          setActiveTab("manual")
        }
      },
    })
  }

  return (
    <div className="space-y-4 pt-2">
      <div className="flex gap-0 border-b-2 border-border">
        <button
          type="button"
          className={`rounded-t-md px-4 py-2 text-sm font-medium transition-colors ${
            activeTab === "manual"
              ? "bg-muted text-foreground"
              : "text-muted-foreground hover:text-foreground hover:bg-muted/50"
          }`}
          onClick={() => setActiveTab("manual")}
        >
          Manual Entry
        </button>
        <button
          type="button"
          className={`rounded-t-md px-4 py-2 text-sm font-medium transition-colors ${
            activeTab === "upload"
              ? "bg-muted text-foreground"
              : "text-muted-foreground hover:text-foreground hover:bg-muted/50"
          }`}
          onClick={() => setActiveTab("upload")}
        >
          File Upload
        </button>
      </div>

      {activeTab === "upload" && (
        <div
          className={`flex flex-col items-center gap-3 rounded-lg border-2 border-dashed p-8 transition-colors ${
            dragActive
              ? "border-primary bg-primary/5"
              : "border-border"
          }`}
          onDragEnter={handleDrag}
          onDragLeave={handleDrag}
          onDragOver={handleDrag}
          onDrop={handleDrop}
        >
          <Upload className={`size-8 ${dragActive ? "text-primary" : "text-muted-foreground"}`} />
          <div className="text-center">
            <p className="text-sm text-muted-foreground">
              Drop your resume (PDF or Markdown) here
            </p>
            <button
              type="button"
              className="mt-1 text-sm font-medium text-primary underline underline-offset-2 hover:text-primary/80"
              onClick={() => fileInputRef.current?.click()}
            >
              or browse files
            </button>
          </div>
          <input
            ref={fileInputRef}
            type="file"
            accept=".pdf,.md,.markdown"
            className="hidden"
            onChange={handleFileSelect}
          />
          {uploadProfile.isPending && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="size-4 animate-spin" />
              Parsing your resume...
            </div>
          )}
          {uploadProfile.isError && (
            <div className="text-center">
              <p className="text-sm text-destructive">
                Could not parse the file — it may be in an unsupported format or corrupted.
              </p>
              <p className="mt-1 text-xs text-muted-foreground">
                Make sure the file is a valid PDF or Markdown. You can also{" "}
                <button
                  type="button"
                  className="text-primary underline underline-offset-2 hover:text-primary/80"
                  onClick={() => setActiveTab("manual")}
                >
                  enter your details manually
                </button>
                .
              </p>
            </div>
          )}
        </div>
      )}

      {activeTab === "manual" && (
        <Form {...form}>
        <form onChange={handleChange} onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <div className="space-y-3">
            <h3 className="text-sm font-semibold text-foreground">Personal Info</h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <FormField
                control={form.control}
                name="fullName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Full Name</FormLabel>
                    <FormControl>
                      <Input placeholder="John Doe" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email</FormLabel>
                    <FormControl>
                      <Input type="email" placeholder="john@example.com" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="phone"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Phone</FormLabel>
                    <FormControl>
                      <Input placeholder="+1 555-0100" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="location"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Location</FormLabel>
                    <FormControl>
                      <Input placeholder="San Francisco, CA" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="linkedinUrl"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>LinkedIn URL</FormLabel>
                    <FormControl>
                      <Input placeholder="https://linkedin.com/in/..." {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="websiteUrl"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Website</FormLabel>
                    <FormControl>
                      <Input placeholder="https://..." {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="githubUrl"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>GitHub URL</FormLabel>
                    <FormControl>
                      <Input placeholder="https://github.com/..." {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
            <FormField
              control={form.control}
              name="professionalSummary"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Professional Summary</FormLabel>
                  <FormControl>
                    <Textarea rows={3} placeholder="Brief overview of your experience..." {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          {/* Work Experience */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-foreground">Work Experience</h3>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() =>
                  workExperienceArray.append({
                    jobTitle: "",
                    company: "",
                    location: "",
                    startDate: "",
                    endDate: "",
                    description: "",
                  })
                }
              >
                <Plus className="mr-1 size-3" />
                Add
              </Button>
            </div>
            {workExperienceArray.fields.length === 0 && (
              <p className="text-sm text-muted-foreground">
                No positions added yet — use the Add button above to include
                roles from your career history.
              </p>
            )}
            {workExperienceArray.fields.map((field, index) => (
              <div key={field.id} className="space-y-2 rounded-md border p-3">
                <div className="flex items-center justify-between">
                  <Badge variant="secondary">Position {index + 1}</Badge>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => workExperienceArray.remove(index)}
                  >
                    <Trash2 className="size-3 text-destructive" />
                  </Button>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <FormField
                    control={form.control}
                    name={`workExperience.${index}.jobTitle`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Job Title</FormLabel>
                        <FormControl>
                          <Input placeholder="Software Engineer" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name={`workExperience.${index}.company`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Company</FormLabel>
                        <FormControl>
                          <Input placeholder="Acme Corp" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name={`workExperience.${index}.location`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Location</FormLabel>
                        <FormControl>
                          <Input placeholder="Remote" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <div className="flex gap-2">
                    <FormField
                      control={form.control}
                      name={`workExperience.${index}.startDate`}
                      render={({ field }) => (
                        <FormItem className="flex-1">
                          <FormLabel>Start Date</FormLabel>
                          <FormControl>
                            <Input placeholder="2020-01" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name={`workExperience.${index}.endDate`}
                      render={({ field }) => (
                        <FormItem className="flex-1">
                          <FormLabel>End Date</FormLabel>
                          <FormControl>
                            <Input placeholder="Present" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                </div>
                <FormField
                  control={form.control}
                  name={`workExperience.${index}.description`}
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Description</FormLabel>
                      <FormControl>
                        <Textarea rows={2} placeholder="Describe your responsibilities..." {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            ))}
          </div>

          {/* Education */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-foreground">Education</h3>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() =>
                  educationArray.append({
                    degree: "",
                    institution: "",
                    location: "",
                    startDate: "",
                    endDate: "",
                    description: "",
                  })
                }
              >
                <Plus className="mr-1 size-3" />
                Add
              </Button>
            </div>
            {educationArray.fields.length === 0 && (
              <p className="text-sm text-muted-foreground">
                No education entries added yet — use the Add button above to
                include your degrees and certifications.
              </p>
            )}
            {educationArray.fields.map((field, index) => (
              <div key={field.id} className="space-y-2 rounded-md border p-3">
                <div className="flex items-center justify-between">
                  <Badge variant="secondary">Education {index + 1}</Badge>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => educationArray.remove(index)}
                  >
                    <Trash2 className="size-3 text-destructive" />
                  </Button>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <FormField
                    control={form.control}
                    name={`education.${index}.degree`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Degree</FormLabel>
                        <FormControl>
                          <Input placeholder="B.S. Computer Science" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name={`education.${index}.institution`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Institution</FormLabel>
                        <FormControl>
                          <Input placeholder="University of..." {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name={`education.${index}.location`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Location</FormLabel>
                        <FormControl>
                          <Input placeholder="City, State" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name={`education.${index}.gpa`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>GPA (optional)</FormLabel>
                        <FormControl>
                          <Input placeholder="3.8" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name={`education.${index}.startDate`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Start Date</FormLabel>
                        <FormControl>
                          <Input placeholder="2016-09" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name={`education.${index}.endDate`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>End Date</FormLabel>
                        <FormControl>
                          <Input placeholder="2020-06" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
                <FormField
                  control={form.control}
                  name={`education.${index}.description`}
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Description</FormLabel>
                      <FormControl>
                        <Textarea rows={2} placeholder="Activities, honors..." {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            ))}
          </div>

          {/* Skills */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-foreground">Skills</h3>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() =>
                  skillsArray.append({ name: "", skills: [] })
                }
              >
                <Plus className="mr-1 size-3" />
                Add Category
              </Button>
            </div>
            {skillsArray.fields.length === 0 && (
              <p className="text-sm text-muted-foreground">
                No skill categories added yet — use Add Category above to group
                your skills into categories like Languages, Frameworks, and Tools.
              </p>
            )}
            {skillsArray.fields.map((field, index) => (
              <div key={field.id} className="space-y-2 rounded-md border p-3">
                <div className="flex items-center justify-between">
                  <FormField
                    control={form.control}
                    name={`skills.${index}.name`}
                    render={({ field }) => (
                      <FormItem className="flex-1">
                        <FormControl>
                          <Input placeholder="Category e.g. Languages" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => skillsArray.remove(index)}
                  >
                    <Trash2 className="size-3 text-destructive" />
                  </Button>
                </div>
                <SkillInput
                  index={index}
                  skills={field.skills}
                  onChange={(skills) => {
                    updateSkills(index, { ...field, skills })
                    handleChange()
                  }}
                />
              </div>
            ))}
          </div>

          {/* Projects */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-foreground">Projects</h3>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() =>
                  projectsArray.append({
                    name: "",
                    role: "",
                    description: "",
                    technologies: [],
                    url: "",
                  })
                }
              >
                <Plus className="mr-1 size-3" />
                Add
              </Button>
            </div>
            {projectsArray.fields.length === 0 && (
              <p className="text-sm text-muted-foreground">
                No projects added yet — use the Add button above to showcase your
                work with project names, descriptions, and technologies.
              </p>
            )}
            {projectsArray.fields.map((field, index) => (
              <div key={field.id} className="space-y-2 rounded-md border p-3">
                <div className="flex items-center justify-between">
                  <Badge variant="secondary">Project {index + 1}</Badge>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => projectsArray.remove(index)}
                  >
                    <Trash2 className="size-3 text-destructive" />
                  </Button>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <FormField
                    control={form.control}
                    name={`projects.${index}.name`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Project Name</FormLabel>
                        <FormControl>
                          <Input placeholder="My Project" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name={`projects.${index}.role`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Role</FormLabel>
                        <FormControl>
                          <Input placeholder="Lead Developer" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
                <FormField
                  control={form.control}
                  name={`projects.${index}.description`}
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Description</FormLabel>
                      <FormControl>
                        <Textarea rows={2} placeholder="Project description..." {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name={`projects.${index}.url`}
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>URL (optional)</FormLabel>
                      <FormControl>
                        <Input placeholder="https://..." {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <div>
                  <Label className="text-sm font-medium">Technologies</Label>
                  <TechInput
                    index={index}
                    technologies={field.technologies}
                    onChange={(technologies) => {
                      updateProjects(index, { ...field, technologies })
                      handleChange()
                    }}
                  />
                </div>
              </div>
            ))}
          </div>

          {/* Certifications */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-foreground">Certifications</h3>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() =>
                  certificationsArray.append({
                    name: "",
                    issuingOrg: "",
                    dateObtained: "",
                  })
                }
              >
                <Plus className="mr-1 size-3" />
                Add
              </Button>
            </div>
            {certificationsArray.fields.length === 0 && (
              <p className="text-sm text-muted-foreground">
                No certifications added yet — use the Add button above to list
                professional certifications that validate your expertise.
              </p>
            )}
            {certificationsArray.fields.map((field, index) => (
              <div key={field.id} className="space-y-2 rounded-md border p-3">
                <div className="flex items-center justify-between">
                  <Badge variant="secondary">Certification {index + 1}</Badge>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => certificationsArray.remove(index)}
                  >
                    <Trash2 className="size-3 text-destructive" />
                  </Button>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <FormField
                    control={form.control}
                    name={`certifications.${index}.name`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Name</FormLabel>
                        <FormControl>
                          <Input placeholder="AWS Solutions Architect" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name={`certifications.${index}.issuingOrg`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Issuing Org</FormLabel>
                        <FormControl>
                          <Input placeholder="Amazon Web Services" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name={`certifications.${index}.dateObtained`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Date Obtained</FormLabel>
                        <FormControl>
                          <Input placeholder="2023-06" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name={`certifications.${index}.expiryDate`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Expiry Date (optional)</FormLabel>
                        <FormControl>
                          <Input placeholder="2026-06" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
                <FormField
                  control={form.control}
                  name={`certifications.${index}.credentialUrl`}
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Credential URL (optional)</FormLabel>
                      <FormControl>
                        <Input placeholder="https://..." {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            ))}
          </div>

          {/* Languages */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-foreground">Languages</h3>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() =>
                  languagesArray.append({
                    name: "",
                    proficiency: "",
                  })
                }
              >
                <Plus className="mr-1 size-3" />
                Add
              </Button>
            </div>
            {languagesArray.fields.length === 0 && (
              <p className="text-sm text-muted-foreground">
                No languages added yet — use the Add button above to include
                languages you speak with their proficiency levels.
              </p>
            )}
            {languagesArray.fields.map((field, index) => (
              <div key={field.id} className="flex items-end gap-2 rounded-md border p-3">
                <FormField
                  control={form.control}
                  name={`languages.${index}.name`}
                  render={({ field }) => (
                    <FormItem className="flex-1">
                      <FormLabel>Language</FormLabel>
                      <FormControl>
                        <Input placeholder="English" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name={`languages.${index}.proficiency`}
                  render={({ field }) => (
                    <FormItem className="flex-1">
                      <FormLabel>Proficiency</FormLabel>
                      <FormControl>
                        <Input placeholder="Native" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => languagesArray.remove(index)}
                >
                  <Trash2 className="size-4 text-destructive" />
                </Button>
              </div>
            ))}
          </div>

          <Button type="submit" disabled={saveProfile.isPending} className="w-full">
            {saveProfile.isPending ? (
              <>
                <Loader2 className="mr-2 size-4 animate-spin" />
                Saving...
              </>
            ) : (
              "Save & Continue"
            )}
          </Button>
        </form>
        </Form>
      )}
    </div>
  )
}

function SkillInput({
  index: _index,
  skills,
  onChange,
}: {
  index: number
  skills: string[]
  onChange: (skills: string[]) => void
}) {
  const [input, setInput] = useState("")

  function addSkill() {
    const trimmed = input.trim()
    if (trimmed && !skills.includes(trimmed)) {
      onChange([...skills, trimmed])
      setInput("")
    }
  }

  function removeSkill(skill: string) {
    onChange(skills.filter((s) => s !== skill))
  }

  return (
    <div className="space-y-1.5">
      <div className="flex gap-1">
        <Input
          placeholder="Type a skill and press Enter"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault()
              addSkill()
            }
          }}
        />
        <Button type="button" variant="outline" size="sm" onClick={addSkill}>
          Add
        </Button>
      </div>
      <div className="flex flex-wrap gap-1">
        {skills.map((skill) => (
          <Badge key={skill} variant="secondary" className="gap-1">
            {skill}
            <button
              type="button"
              onClick={() => removeSkill(skill)}
              className="ml-0.5 text-muted-foreground hover:text-destructive"
            >
              <Trash2 className="size-3" />
            </button>
          </Badge>
        ))}
      </div>
    </div>
  )
}

function TechInput({
  index: _index,
  technologies,
  onChange,
}: {
  index: number
  technologies: string[]
  onChange: (technologies: string[]) => void
}) {
  const [input, setInput] = useState("")

  function addTech() {
    const trimmed = input.trim()
    if (trimmed && !technologies.includes(trimmed)) {
      onChange([...technologies, trimmed])
      setInput("")
    }
  }

  function removeTech(tech: string) {
    onChange(technologies.filter((t) => t !== tech))
  }

  return (
    <div className="space-y-1.5">
      <div className="flex gap-1">
        <Input
          placeholder="Type a technology and press Enter"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault()
              addTech()
            }
          }}
        />
        <Button type="button" variant="outline" size="sm" onClick={addTech}>
          Add
        </Button>
      </div>
      <div className="flex flex-wrap gap-1">
        {technologies.map((tech) => (
          <Badge key={tech} variant="secondary" className="gap-1">
            {tech}
            <button
              type="button"
              onClick={() => removeTech(tech)}
              className="ml-0.5 text-muted-foreground hover:text-destructive"
            >
              <Trash2 className="size-3" />
            </button>
          </Badge>
        ))}
      </div>
    </div>
  )
}

export { ProfileSection }
export type { ProfileSectionProps }
