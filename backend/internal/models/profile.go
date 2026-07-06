package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type WorkExperience struct {
	JobTitle    string `json:"jobTitle"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Description string `json:"description"`
}

type Education struct {
	Degree      string `json:"degree"`
	Institution string `json:"institution"`
	Location    string `json:"location"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	GPA         string `json:"gpa"`
}

type SkillCategory struct {
	Name   string   `json:"name"`
	Skills []string `json:"skills"`
}

type Project struct {
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	Description  string   `json:"description"`
	Technologies []string `json:"technologies"`
	URL          string   `json:"url"`
}

type Certification struct {
	Name          string `json:"name"`
	IssuingOrg    string `json:"issuingOrg"`
	DateObtained  string `json:"dateObtained"`
	ExpiryDate    string `json:"expiryDate"`
	CredentialURL string `json:"credentialUrl"`
}

type Language struct {
	Name        string `json:"name"`
	Proficiency string `json:"proficiency"`
}

type WorkExperienceList []WorkExperience

func (l *WorkExperienceList) Scan(src any) error {
	return scanJSON(l, src)
}

func (l WorkExperienceList) Value() (driver.Value, error) {
	return valueJSON(l)
}

type EducationList []Education

func (l *EducationList) Scan(src any) error {
	return scanJSON(l, src)
}

func (l EducationList) Value() (driver.Value, error) {
	return valueJSON(l)
}

type SkillCategoryList []SkillCategory

func (l *SkillCategoryList) Scan(src any) error {
	return scanJSON(l, src)
}

func (l SkillCategoryList) Value() (driver.Value, error) {
	return valueJSON(l)
}

type ProjectList []Project

func (l *ProjectList) Scan(src any) error {
	return scanJSON(l, src)
}

func (l ProjectList) Value() (driver.Value, error) {
	return valueJSON(l)
}

type CertificationList []Certification

func (l *CertificationList) Scan(src any) error {
	return scanJSON(l, src)
}

func (l CertificationList) Value() (driver.Value, error) {
	return valueJSON(l)
}

type LanguageList []Language

func (l *LanguageList) Scan(src any) error {
	return scanJSON(l, src)
}

func (l LanguageList) Value() (driver.Value, error) {
	return valueJSON(l)
}

type Profile struct {
	ID                  string             `json:"id"`
	FullName            string             `json:"fullName"`
	Email               string             `json:"email"`
	Phone               string             `json:"phone"`
	Location            string             `json:"location"`
	LinkedInURL         string             `json:"linkedinUrl"`
	WebsiteURL          string             `json:"websiteUrl"`
	GitHubURL           string             `json:"githubUrl"`
	ProfessionalSummary string             `json:"professionalSummary"`
	WorkExperience      WorkExperienceList `json:"workExperience"`
	Education           EducationList      `json:"education"`
	Skills              SkillCategoryList  `json:"skills"`
	Projects            ProjectList        `json:"projects"`
	Certifications      CertificationList  `json:"certifications"`
	Languages           LanguageList       `json:"languages"`
	CreatedAt           time.Time          `json:"createdAt"`
	UpdatedAt           time.Time          `json:"updatedAt"`
}

func scanJSON(dest any, src any) error {
	if src == nil {
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("models: unsupported scan source type %T", src)
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, dest)
}

func valueJSON(v any) (driver.Value, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("models: marshal: %w", err)
	}
	return string(data), nil
}
