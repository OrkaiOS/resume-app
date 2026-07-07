package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/marco/resume-app/internal/llm"
	"github.com/marco/resume-app/internal/models"
)

type ProfileParser struct{}

func NewProfileParser() *ProfileParser {
	return &ProfileParser{}
}

type ProfileLLMParser struct{}

func NewProfileLLMParser() *ProfileLLMParser {
	return &ProfileLLMParser{}
}

func (p *ProfileParser) ParsePDF(r io.Reader) (*models.Profile, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("services.ProfileParser.ParsePDF: read: %w", err)
	}

	pdfReader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("services.ProfileParser.ParsePDF: open pdf: %w", err)
	}

	textReader, err := pdfReader.GetPlainText()
	if err != nil {
		return nil, fmt.Errorf("services.ProfileParser.ParsePDF: extract text: %w", err)
	}

	text, err := io.ReadAll(textReader)
	if err != nil {
		return nil, fmt.Errorf("services.ProfileParser.ParsePDF: read text: %w", err)
	}

	profile := &models.Profile{
		ProfessionalSummary: strings.TrimSpace(string(text)),
	}

	return profile, nil
}

func (p *ProfileParser) ParseMarkdown(r io.Reader) (*models.Profile, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("services.ProfileParser.ParseMarkdown: %w", err)
	}

	text := string(data)
	profile := &models.Profile{}

	lines := strings.Split(text, "\n")
	bodyStart := 0

	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "# ") {
		profile.FullName = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[0]), "# "))
		bodyStart = 1
	}

	sections := splitSections(strings.Join(lines[bodyStart:], "\n"))
	for heading, content := range sections {
		parseSection(profile, heading, content)
	}

	return profile, nil
}

func splitSections(text string) map[string]string {
	sections := make(map[string]string)
	parts := strings.Split(text, "\n## ")

	if len(parts) == 0 {
		return sections
	}

	if !strings.HasPrefix(text, "## ") {
		parts = parts[1:]
	}

	for _, part := range parts {
		idx := strings.Index(part, "\n")
		if idx == -1 {
			sections[strings.TrimSpace(part)] = ""
		} else {
			sections[strings.TrimSpace(part[:idx])] = strings.TrimSpace(part[idx+1:])
		}
	}

	return sections
}

func parseSection(profile *models.Profile, heading, content string) {
	lower := strings.ToLower(heading)

	switch {
	case contains(lower, "contact"):
		parseContact(profile, content)
	case contains(lower, "summary") || contains(lower, "professional"):
		profile.ProfessionalSummary = strings.TrimSpace(content)
	case contains(lower, "experience") || contains(lower, "work"):
		profile.WorkExperience = parseWorkExperience(content)
	case contains(lower, "education"):
		profile.Education = parseEducation(content)
	case contains(lower, "skill"):
		profile.Skills = parseSkillCategories(content)
	case contains(lower, "project"):
		profile.Projects = parseProjects(content)
	case contains(lower, "certification"):
		profile.Certifications = parseCertifications(content)
	case contains(lower, "language"):
		profile.Languages = parseLanguages(content)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func parseContact(profile *models.Profile, content string) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(strings.ToLower(parts[0]))
		val := strings.TrimSpace(parts[1])

		switch key {
		case "email":
			profile.Email = val
		case "phone":
			profile.Phone = val
		case "location":
			profile.Location = val
		case "linkedin":
			profile.LinkedInURL = val
		case "website":
			profile.WebsiteURL = val
		case "github":
			profile.GitHubURL = val
		}
	}
}

func parseWorkExperience(content string) models.WorkExperienceList {
	var entries []models.WorkExperience
	blocks := strings.Split(content, "\n### ")

	if len(blocks) > 1 {
		blocks = blocks[1:]
		for _, block := range blocks {
			entry := parseWorkEntry(block)
			entries = append(entries, entry)
		}
	} else {
		lines := strings.Split(strings.TrimSpace(content), "\n")
		if len(lines) > 0 {
			entry := parseWorkEntryFromBullets(lines)
			entries = append(entries, entry)
		}
	}

	var result models.WorkExperienceList
	for _, e := range entries {
		result = append(result, e)
	}
	return result
}

func parseWorkEntry(block string) models.WorkExperience {
	lines := strings.Split(strings.TrimSpace(block), "\n")
	if len(lines) == 0 {
		return models.WorkExperience{}
	}

	entry := models.WorkExperience{}

	firstLine := strings.TrimSpace(lines[0])
	if idx := strings.LastIndex(firstLine, " at "); idx >= 0 {
		entry.JobTitle = strings.TrimSpace(firstLine[:idx])
		entry.Company = strings.TrimSpace(firstLine[idx+4:])
	} else {
		entry.JobTitle = firstLine
	}

	descStart := 1
	if len(lines) > 1 {
		meta := strings.TrimSpace(lines[1])
		if strings.Contains(meta, "|") {
			metaParts := strings.SplitN(meta, "|", 2)
			entry.Location = strings.TrimSpace(metaParts[0])
			if len(metaParts) > 1 {
				dates := strings.TrimSpace(metaParts[1])
				if dashIdx := strings.Index(dates, " - "); dashIdx >= 0 {
					entry.StartDate = strings.TrimSpace(dates[:dashIdx])
					entry.EndDate = strings.TrimSpace(dates[dashIdx+3:])
				}
			}
			descStart = 2
		}
	}

	var descLines []string
	for _, line := range lines[descStart:] {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		if line != "" {
			descLines = append(descLines, line)
		}
	}
	entry.Description = strings.Join(descLines, "\n")

	return entry
}

func parseWorkEntryFromBullets(lines []string) models.WorkExperience {
	entry := models.WorkExperience{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			entry.Description += line + "\n"
			continue
		}
		key := strings.TrimSpace(strings.ToLower(parts[0]))
		val := strings.TrimSpace(parts[1])
		switch key {
		case "title", "job title", "role":
			entry.JobTitle = val
		case "company":
			entry.Company = val
		case "location":
			entry.Location = val
		case "start date", "start":
			entry.StartDate = val
		case "end date", "end":
			entry.EndDate = val
		default:
			entry.Description += line + "\n"
		}
	}
	entry.Description = strings.TrimSpace(entry.Description)
	return entry
}

func parseEducation(content string) models.EducationList {
	var entries []models.Education
	blocks := strings.Split(content, "\n### ")

	if len(blocks) > 1 {
		blocks = blocks[1:]
	}
	for _, block := range blocks {
		entry := parseEducationEntry(block)
		entries = append(entries, entry)
	}

	var result models.EducationList
	for _, e := range entries {
		result = append(result, e)
	}
	return result
}

func parseEducationEntry(block string) models.Education {
	lines := strings.Split(strings.TrimSpace(block), "\n")
	if len(lines) == 0 {
		return models.Education{}
	}

	entry := models.Education{Degree: strings.TrimSpace(lines[0])}

	descStart := 1
	if len(lines) > 1 {
		meta := strings.TrimSpace(lines[1])
		if strings.Contains(meta, "|") {
			metaParts := strings.SplitN(meta, "|", 2)
			entry.Institution = strings.TrimSpace(metaParts[0])
			if len(metaParts) > 1 {
				dates := strings.TrimSpace(metaParts[1])
				if dashIdx := strings.Index(dates, " - "); dashIdx >= 0 {
					entry.StartDate = strings.TrimSpace(dates[:dashIdx])
					entry.EndDate = strings.TrimSpace(dates[dashIdx+3:])
				}
			}
			descStart = 2
		}
	}

	for _, line := range lines[descStart:] {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), "gpa") {
			entry.GPA = strings.TrimSpace(parts[1])
		}
	}

	return entry
}

func parseSkillCategories(content string) models.SkillCategoryList {
	var categories []models.SkillCategory

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		skillsStr := strings.TrimSpace(parts[1])
		var skills []string
		for _, s := range strings.Split(skillsStr, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				skills = append(skills, s)
			}
		}

		if len(skills) > 0 {
			categories = append(categories, models.SkillCategory{Name: name, Skills: skills})
		}
	}

	if len(categories) == 0 {
		var skills []string
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimPrefix(line, "* ")
			if line != "" {
				skills = append(skills, line)
			}
		}
		if len(skills) > 0 {
			categories = append(categories, models.SkillCategory{Name: "Skills", Skills: skills})
		}
	}

	var result models.SkillCategoryList
	for _, c := range categories {
		result = append(result, c)
	}
	return result
}

func parseProjects(content string) models.ProjectList {
	var projects []models.Project
	blocks := strings.Split(content, "\n### ")

	if len(blocks) > 1 {
		blocks = blocks[1:]
	}
	for _, block := range blocks {
		projects = append(projects, parseProjectEntry(block))
	}

	var result models.ProjectList
	for _, p := range projects {
		result = append(result, p)
	}
	return result
}

func parseProjectEntry(block string) models.Project {
	lines := strings.Split(strings.TrimSpace(block), "\n")
	if len(lines) == 0 {
		return models.Project{}
	}

	entry := models.Project{Name: strings.TrimSpace(lines[0])}

	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			entry.Description += line + "\n"
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		val := strings.TrimSpace(parts[1])

		switch key {
		case "role":
			entry.Role = val
		case "description":
			entry.Description = val
		case "technologies", "tech":
			for _, t := range strings.Split(val, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					entry.Technologies = append(entry.Technologies, t)
				}
			}
		case "url", "link":
			entry.URL = val
		default:
			entry.Description += line + "\n"
		}
	}
	entry.Description = strings.TrimSpace(entry.Description)

	return entry
}

func parseCertifications(content string) models.CertificationList {
	var certs []models.Certification

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " - ", 2)
		if len(parts) == 2 {
			certs = append(certs, models.Certification{
				Name:       strings.TrimSpace(parts[0]),
				IssuingOrg: strings.TrimSpace(parts[1]),
			})
		} else {
			certs = append(certs, models.Certification{Name: line})
		}
	}

	var result models.CertificationList
	for _, c := range certs {
		result = append(result, c)
	}
	return result
}

func parseLanguages(content string) models.LanguageList {
	var langs []models.Language

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "(", 2)
		if len(parts) == 2 {
			langs = append(langs, models.Language{
				Name:        strings.TrimSpace(parts[0]),
				Proficiency: strings.TrimSpace(strings.TrimSuffix(parts[1], ")")),
			})
		} else {
			langs = append(langs, models.Language{Name: line})
		}
	}

	var result models.LanguageList
	for _, l := range langs {
		result = append(result, l)
	}
	return result
}

const profileExtractionSystemPrompt = `You are a resume data extraction tool. Extract structured profile data from the provided resume text. Return ONLY a JSON object (no markdown fences, no explanations). Use the exact field names below.

{
  "fullName": "string",
  "email": "string",
  "phone": "string",
  "location": "string",
  "linkedinUrl": "string",
  "websiteUrl": "string",
  "githubUrl": "string",
  "professionalSummary": "string",
  "workExperience": [
    {
      "jobTitle": "string",
      "company": "string",
      "location": "string",
      "startDate": "string (YYYY-MM or YYYY)",
      "endDate": "string (YYYY-MM, YYYY, or 'Present')",
      "description": "string"
    }
  ],
  "education": [
    {
      "degree": "string",
      "institution": "string",
      "location": "string",
      "startDate": "string",
      "endDate": "string",
      "gpa": "string"
    }
  ],
  "skills": [
    {
      "name": "string (category, e.g. 'Languages', 'Frameworks', 'Tools')",
      "skills": ["string"]
    }
  ],
  "projects": [
    {
      "name": "string",
      "role": "string",
      "description": "string",
      "technologies": ["string"],
      "url": "string"
    }
  ],
  "certifications": [
    {
      "name": "string",
      "issuingOrg": "string",
      "dateObtained": "string",
      "expiryDate": "string",
      "credentialUrl": "string"
    }
  ],
  "languages": [
    {
      "name": "string (e.g. 'English', 'Spanish')",
      "proficiency": "string (e.g. 'Native', 'Fluent', 'Intermediate')"
    }
  ]
}

Rules:
- Only include fields that appear in the resume. Omit missing fields entirely.
- Keep descriptions concise but complete.
- Group skills into logical categories (Languages, Frameworks, Tools, Platforms, etc.).
- If a section is absent, use an empty array [] for that field.
- Return valid JSON only, no markdown fences.`

func (p *ProfileLLMParser) Parse(ctx context.Context, text string, provider, model, apiKey string) (*models.Profile, error) {
	if provider == "" || model == "" {
		return nil, fmt.Errorf("services.ProfileLLMParser.Parse: LLM not configured (provider=%q model=%q)", provider, model)
	}

	client := llm.NewClient(provider, model, apiKey)
	response, err := client.Chat(ctx, profileExtractionSystemPrompt, text)
	if err != nil {
		return nil, fmt.Errorf("services.ProfileLLMParser.Parse: llm call: %w", err)
	}

	jsonText := extractJSON(response)

	var profile models.Profile
	if err := json.Unmarshal([]byte(jsonText), &profile); err != nil {
		return nil, fmt.Errorf("services.ProfileLLMParser.Parse: unmarshal JSON: %w", err)
	}

	return &profile, nil
}

func extractJSON(text string) string {
	text = strings.TrimSpace(text)
	if idx := strings.Index(text, "```json"); idx >= 0 {
		start := idx + len("```json")
		if end := strings.Index(text[start:], "```"); end >= 0 {
			return strings.TrimSpace(text[start : start+end])
		}
	}
	if idx := strings.Index(text, "```"); idx >= 0 {
		start := idx + len("```")
		if end := strings.Index(text[start:], "```"); end >= 0 {
			candidate := strings.TrimSpace(text[start : start+end])
			if strings.HasPrefix(candidate, "{") {
				return candidate
			}
		}
	}
	if idx := strings.Index(text, "{"); idx >= 0 {
		lastBrace := strings.LastIndex(text, "}")
		if lastBrace > idx {
			return text[idx : lastBrace+1]
		}
	}
	return text
}
