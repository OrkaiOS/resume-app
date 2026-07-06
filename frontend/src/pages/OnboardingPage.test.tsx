import { describe, it, expect } from "vitest"
import {
  llmConfigSchema,
  profileSchema,
  workExperienceSchema,
  educationSchema,
  projectSchema,
  certificationSchema,
  languageSchema,
  skillCategorySchema,
} from "@/features/onboarding/schemas"

describe("llmConfigSchema", () => {
  it("accepts valid ollama config", () => {
    const result = llmConfigSchema.safeParse({
      provider: "ollama",
      apiKey: "",
      model: "llama3",
    })
    expect(result.success).toBe(true)
  })

  it("accepts valid openai config", () => {
    const result = llmConfigSchema.safeParse({
      provider: "openai",
      apiKey: "sk-abc123",
      model: "gpt-4",
    })
    expect(result.success).toBe(true)
  })

  it("rejects invalid provider", () => {
    const result = llmConfigSchema.safeParse({
      provider: "invalid",
    })
    expect(result.success).toBe(false)
  })

  it("defaults optional fields", () => {
    const result = llmConfigSchema.safeParse({
      provider: "anthropic",
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.apiKey).toBe("")
      expect(result.data.model).toBe("")
    }
  })
})

describe("profileSchema", () => {
  it("accepts minimal valid profile", () => {
    const result = profileSchema.safeParse({
      fullName: "John Doe",
      email: "john@example.com",
    })
    expect(result.success).toBe(true)
  })

  it("rejects missing fullName", () => {
    const result = profileSchema.safeParse({
      email: "john@example.com",
    })
    expect(result.success).toBe(false)
    if (!result.success) {
      const msgs = result.error.issues.map((i) => i.path.join("."))
      expect(msgs).toContain("fullName")
    }
  })

  it("rejects invalid email", () => {
    const result = profileSchema.safeParse({
      fullName: "John Doe",
      email: "not-an-email",
    })
    expect(result.success).toBe(false)
  })

  it("rejects invalid URL in linkedinUrl", () => {
    const result = profileSchema.safeParse({
      fullName: "John Doe",
      email: "john@example.com",
      linkedinUrl: "not-a-url",
    })
    expect(result.success).toBe(false)
  })

  it("accepts complete profile with all fields", () => {
    const result = profileSchema.safeParse({
      fullName: "Jane Smith",
      email: "jane@example.com",
      phone: "+1 555-0100",
      location: "San Francisco, CA",
      linkedinUrl: "https://linkedin.com/in/janesmith",
      websiteUrl: "https://janesmith.dev",
      githubUrl: "https://github.com/janesmith",
      professionalSummary: "Experienced software engineer",
      workExperience: [
        {
          jobTitle: "Senior Engineer",
          company: "Acme Corp",
          location: "Remote",
          startDate: "2022-01",
          endDate: "Present",
          description: "Built things",
        },
      ],
      education: [
        {
          degree: "B.S. Computer Science",
          institution: "Stanford",
          location: "Stanford, CA",
          startDate: "2016-09",
          endDate: "2020-06",
          gpa: "3.8",
          description: "Honors",
        },
      ],
      skills: [
        { name: "Languages", skills: ["TypeScript", "Go"] },
        { name: "Frameworks", skills: ["React", "Gin"] },
      ],
      projects: [
        {
          name: "My Project",
          role: "Lead",
          description: "A cool project",
          technologies: ["React", "Go"],
          url: "https://github.com/user/project",
        },
      ],
      certifications: [
        {
          name: "AWS Solutions Architect",
          issuingOrg: "AWS",
          dateObtained: "2024-01",
          expiryDate: "2027-01",
          credentialUrl: "https://aws.com/cert/123",
        },
      ],
      languages: [
        { name: "English", proficiency: "Native" },
        { name: "Spanish", proficiency: "Intermediate" },
      ],
    })
    expect(result.success).toBe(true)
  })

  it("defaults empty arrays and optional strings", () => {
    const result = profileSchema.safeParse({
      fullName: "John Doe",
      email: "john@example.com",
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.workExperience).toEqual([])
      expect(result.data.education).toEqual([])
      expect(result.data.skills).toEqual([])
      expect(result.data.projects).toEqual([])
      expect(result.data.certifications).toEqual([])
      expect(result.data.languages).toEqual([])
      expect(result.data.phone).toBe("")
    }
  })
})

describe("workExperienceSchema", () => {
  it("accepts valid work experience", () => {
    const result = workExperienceSchema.safeParse({
      jobTitle: "Engineer",
      company: "Corp",
      location: "Remote",
      startDate: "2020-01",
      endDate: "2023-01",
    })
    expect(result.success).toBe(true)
  })

  it("rejects missing required fields", () => {
    const result = workExperienceSchema.safeParse({
      jobTitle: "Engineer",
    })
    expect(result.success).toBe(false)
  })
})

describe("educationSchema", () => {
  it("accepts valid education", () => {
    const result = educationSchema.safeParse({
      degree: "B.S.",
      institution: "University",
      location: "City",
      startDate: "2016-09",
      endDate: "2020-06",
    })
    expect(result.success).toBe(true)
  })

  it("rejects missing degree", () => {
    const result = educationSchema.safeParse({
      institution: "University",
      location: "City",
      startDate: "2016-09",
      endDate: "2020-06",
    })
    expect(result.success).toBe(false)
  })
})

describe("projectSchema", () => {
  it("accepts valid project with URL", () => {
    const result = projectSchema.safeParse({
      name: "My Project",
      role: "Dev",
      description: "A project",
      technologies: ["Go"],
      url: "https://github.com",
    })
    expect(result.success).toBe(true)
  })

  it("accepts project without URL", () => {
    const result = projectSchema.safeParse({
      name: "My Project",
      role: "Dev",
      description: "A project",
      technologies: ["Go"],
    })
    expect(result.success).toBe(true)
  })

  it("rejects empty technologies array", () => {
    const result = projectSchema.safeParse({
      name: "My Project",
      role: "Dev",
      description: "A project",
      technologies: [],
    })
    expect(result.success).toBe(false)
  })
})

describe("certificationSchema", () => {
  it("accepts valid certification", () => {
    const result = certificationSchema.safeParse({
      name: "AWS Cert",
      issuingOrg: "AWS",
      dateObtained: "2024-01",
    })
    expect(result.success).toBe(true)
  })

  it("rejects missing name", () => {
    const result = certificationSchema.safeParse({
      issuingOrg: "AWS",
      dateObtained: "2024-01",
    })
    expect(result.success).toBe(false)
  })
})

describe("languageSchema", () => {
  it("accepts valid language", () => {
    const result = languageSchema.safeParse({
      name: "English",
      proficiency: "Native",
    })
    expect(result.success).toBe(true)
  })

  it("rejects missing proficiency", () => {
    const result = languageSchema.safeParse({
      name: "English",
    })
    expect(result.success).toBe(false)
  })
})

describe("skillCategorySchema", () => {
  it("accepts valid skill category", () => {
    const result = skillCategorySchema.safeParse({
      name: "Languages",
      skills: ["TypeScript", "Go"],
    })
    expect(result.success).toBe(true)
  })

  it("rejects empty skills array", () => {
    const result = skillCategorySchema.safeParse({
      name: "Languages",
      skills: [],
    })
    expect(result.success).toBe(false)
  })
})
