package rules

import "github.com/singh-anurag-7991/cloud-guard/internal/models"

type Rule interface {
	Name() string
	Evaluate(resource models.Resource) *models.Finding
}

type Engine struct {
	Rules []Rule
}

func NewEngine(rules []Rule) *Engine {
	return &Engine{Rules: rules}
}

func (e *Engine) Evaluate(resources []models.Resource) []models.Finding {
	var findings []models.Finding
	for _, res := range resources {
		for _, rule := range e.Rules {
			if finding := rule.Evaluate(res); finding != nil {
				findings = append(findings, *finding)
			}
		}
	}
	return findings
}
