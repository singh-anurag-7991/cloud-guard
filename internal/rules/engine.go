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
			finding := rule.Evaluate(res)
			if finding == nil {
				continue
			}

			// Backfill fields the engine always knows, so individual rules
			// cannot forget them. Now that scans cover every enabled region,
			// a finding without a region tells the customer to go hunting for
			// a resource ID across 18 consoles.
			if finding.Region == "" {
				finding.Region = res.Region
			}
			if finding.RuleID == "" {
				finding.RuleID = rule.Name()
			}

			findings = append(findings, *finding)
		}
	}
	return findings
}
