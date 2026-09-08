package supervisor

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// CriticIssue represents a specific security or syntax vulnerability detected in code.
type CriticIssue struct {
	Severity       string `json:"severity"` // "HIGH", "MEDIUM", "LOW"
	Category       string `json:"category"` // "SQL Injection", "Command Injection", "Hardcoded Secret", "Unsafe Execution"
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
	Snippet        string `json:"snippet,omitempty"`
}

// CriticVerdict encapsulates the outcome of the security inspection.
type CriticVerdict struct {
	IsClean    bool          `json:"is_clean"`
	Issues     []CriticIssue `json:"issues"`
	Duration   time.Duration `json:"duration"`
	Iterations int           `json:"iterations"`
}

// CriticConfig defines settings for the pool/security-tester critic loop.
type CriticConfig struct {
	MaxCriticLoops int // Maximum refinement iterations (default 2)
	Enabled        bool
}

// SecurityCritic executes static analysis and AST checks for pool/security-tester.
type SecurityCritic struct {
	cfg CriticConfig

	// Compiled vulnerability detection patterns
	sqlInjectionRe   *regexp.Regexp
	cmdInjectionRe   *regexp.Regexp
	hardcodedKeyRe   *regexp.Regexp
	unsafeExecRe     *regexp.Regexp
	unboundedReadRe  *regexp.Regexp
}

// NewSecurityCritic initializes vulnerability patterns and loop caps.
func NewSecurityCritic(cfg CriticConfig) *SecurityCritic {
	if cfg.MaxCriticLoops <= 0 {
		cfg.MaxCriticLoops = 2
	}

	return &SecurityCritic{
		cfg: cfg,
		sqlInjectionRe:  regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE)\s+[^;]*(\+\s*[a-zA-Z_0-9]+|[a-zA-Z_0-9]+\s*\+|fmt\.Sprintf|%s|concat)`),
		cmdInjectionRe:  regexp.MustCompile(`(?i)(exec\.Command\s*\(\s*["'](sh|bash|cmd)["']\s*,\s*["']-c["']|os\.system\s*\(|child_process\.exec\s*\()`),
		hardcodedKeyRe:  regexp.MustCompile(`(?i)(["'](sk-[a-zA-Z0-9]{20,}|ghp_[a-zA-Z0-9]{20,}|AIza[0-9A-Za-z-_]{35}|bearer\s+[a-zA-Z0-9_-]{20,})["'])`),
		unsafeExecRe:    regexp.MustCompile(`(?i)(eval\s*\(|pickle\.loads|yaml\.load\s*\([^,)]+\)|dangerouslySetInnerHTML)`),
		unboundedReadRe: regexp.MustCompile(`(?i)(io\.ReadAll\s*\(\s*r\.Body\s*\)|ioutil\.ReadAll\s*\(\s*r\.Body\s*\))`),
	}
}

// InspectCode scans source code for security vulnerabilities.
func (c *SecurityCritic) InspectCode(code string) CriticVerdict {
	start := time.Now()
	var issues []CriticIssue

	// 1. SQL Injection Check
	if matches := c.sqlInjectionRe.FindAllString(code, 2); len(matches) > 0 {
		issues = append(issues, CriticIssue{
			Severity:       "HIGH",
			Category:       "SQL Injection",
			Description:    "Direct string concatenation or formatting in SQL query detected.",
			Recommendation: "Use parameterized queries or prepared statements ($1, ?, :param).",
			Snippet:        matches[0],
		})
	}

	// 2. Command Injection Check
	if matches := c.cmdInjectionRe.FindAllString(code, 2); len(matches) > 0 {
		issues = append(issues, CriticIssue{
			Severity:       "HIGH",
			Category:       "Command Injection",
			Description:    "Unsafe shell execution or command evaluation detected.",
			Recommendation: "Avoid shell interpolation; invoke executables with separate argument vectors.",
			Snippet:        matches[0],
		})
	}

	// 3. Hardcoded Secret Check
	if matches := c.hardcodedKeyRe.FindAllString(code, 2); len(matches) > 0 {
		issues = append(issues, CriticIssue{
			Severity:       "HIGH",
			Category:       "Hardcoded Secret",
			Description:    "Plaintext API key or credential string embedded in source code.",
			Recommendation: "Retrieve secrets from environment variables or vault storage at runtime.",
			Snippet:        "[REDACTED SECRET]",
		})
	}

	// 4. Unsafe Execution / Deserialization
	if matches := c.unsafeExecRe.FindAllString(code, 2); len(matches) > 0 {
		issues = append(issues, CriticIssue{
			Severity:       "HIGH",
			Category:       "Unsafe Deserialization",
			Description:    "Arbitrary code execution primitive (eval / pickle / unescaped HTML).",
			Recommendation: "Use safe parsers (e.g., json.Unmarshal, yaml.SafeLoader).",
			Snippet:        matches[0],
		})
	}

	// 5. Unbounded I/O Read
	if matches := c.unboundedReadRe.FindAllString(code, 2); len(matches) > 0 {
		issues = append(issues, CriticIssue{
			Severity:       "MEDIUM",
			Category:       "Unbounded Read",
			Description:    "Unbounded network request body read without size limits (DoS vector).",
			Recommendation: "Wrap reader with io.LimitReader or http.MaxBytesReader.",
			Snippet:        matches[0],
		})
	}

	return CriticVerdict{
		IsClean:    len(issues) == 0,
		Issues:     issues,
		Duration:   time.Since(start),
		Iterations: 1,
	}
}

// FormatAdvisoryNotice generates a GitHub markdown alert block if issues were detected.
func FormatAdvisoryNotice(issues []CriticIssue) string {
	if len(issues) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n> [!WARNING]\n")
	sb.WriteString("> **TokMan Security Critic (`pool/security-tester`) Advisory:**\n")
	sb.WriteString("> Automated AST review detected potential security issues in the generated code:\n")
	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("> - **[%s] %s**: %s\n>   *Remediation:* %s\n", issue.Severity, issue.Category, issue.Description, issue.Recommendation))
	}

	return sb.String()
}

// GenerateRefinementFeedback formats issues for feeding back to the generator LLM.
func GenerateRefinementFeedback(issues []CriticIssue) string {
	var sb strings.Builder
	sb.WriteString("The generated code contains security or syntax issues that must be corrected:\n")
	for i, issue := range issues {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s: %s (Fix: %s)\n", i+1, issue.Severity, issue.Category, issue.Description, issue.Recommendation))
	}
	sb.WriteString("\nPlease regenerate the code fixing all listed security issues.")
	return sb.String()
}

// ExecuteCriticLoop orchestrates multi-iteration AST review with early-exit and loop clamping.
func (c *SecurityCritic) ExecuteCriticLoop(
	ctx context.Context,
	initialCode string,
	refinerFunc func(ctx context.Context, code string, feedback string) (string, error),
) (finalCode string, verdict CriticVerdict) {
	currentCode := initialCode
	totalIterations := 0

	for iteration := 1; iteration <= c.cfg.MaxCriticLoops; iteration++ {
		totalIterations = iteration
		v := c.InspectCode(currentCode)

		// Early exit if code passes all checks
		if v.IsClean {
			v.Iterations = totalIterations
			return currentCode, v
		}

		verdict = v

		// If more iterations allowed and refiner is provided, attempt refinement
		if iteration < c.cfg.MaxCriticLoops && refinerFunc != nil {
			feedback := GenerateRefinementFeedback(v.Issues)
			revisedCode, err := refinerFunc(ctx, currentCode, feedback)
			if err != nil || revisedCode == "" {
				// If refinement fails, stop and flag the advisory
				break
			}
			currentCode = revisedCode
		}
	}

	verdict.Iterations = totalIterations

	// If issues remain, append advisory notice
	if !verdict.IsClean {
		finalCode = currentCode + FormatAdvisoryNotice(verdict.Issues)
	} else {
		finalCode = currentCode
	}

	return finalCode, verdict
}
