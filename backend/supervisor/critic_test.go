package supervisor

import (
	"context"
	"strings"
	"testing"
)

func TestSecurityCritic_CleanCode(t *testing.T) {
	critic := NewSecurityCritic(CriticConfig{MaxCriticLoops: 2})
	cleanCode := `
func add(a, b int) int {
	return a + b
}
`
	verdict := critic.InspectCode(cleanCode)
	if !verdict.IsClean {
		t.Fatalf("expected clean code verdict, got %d issues: %+v", len(verdict.Issues), verdict.Issues)
	}
}

func TestSecurityCritic_VulnerabilitiesDetected(t *testing.T) {
	critic := NewSecurityCritic(CriticConfig{MaxCriticLoops: 2})

	cases := []struct {
		name             string
		code             string
		expectedCategory string
	}{
		{
			name:             "SQL Injection via concatenation",
			code:             `query := "SELECT * FROM users WHERE id = '" + userInput + "'"` ,
			expectedCategory: "SQL Injection",
		},
		{
			name:             "Command Injection via shell exec",
			code:             `cmd := exec.Command("sh", "-c", userCommand)`,
			expectedCategory: "Command Injection",
		},
		{
			name:             "Hardcoded Secret",
			code:             `apiKey := "sk-1234567890abcdef1234567890abcdef"`,
			expectedCategory: "Hardcoded Secret",
		},
		{
			name:             "Unsafe eval execution",
			code:             `result := eval(untrustedInput)`,
			expectedCategory: "Unsafe Deserialization",
		},
		{
			name:             "Unbounded network read",
			code:             `data, err := io.ReadAll(r.Body)`,
			expectedCategory: "Unbounded Read",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verdict := critic.InspectCode(tc.code)
			if verdict.IsClean {
				t.Fatalf("expected vulnerability in %s, got clean verdict", tc.name)
			}
			found := false
			for _, issue := range verdict.Issues {
				if issue.Category == tc.expectedCategory {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected category %s, got issues: %+v", tc.expectedCategory, verdict.Issues)
			}
		})
	}
}

func TestSecurityCritic_MultiIterationLoopAndClamping(t *testing.T) {
	critic := NewSecurityCritic(CriticConfig{MaxCriticLoops: 2})
	ctx := context.Background()

	// 1. Successful refinement on iteration 2
	refinementCalls := 0
	badCode := `query := "SELECT * FROM users WHERE id = '" + id + "'"`
	cleanCode := `query := "SELECT * FROM users WHERE id = $1"`

	finalCode, verdict := critic.ExecuteCriticLoop(ctx, badCode, func(ctx context.Context, code string, feedback string) (string, error) {
		refinementCalls++
		return cleanCode, nil
	})

	if !verdict.IsClean {
		t.Errorf("expected clean verdict after refinement, got issues: %+v", verdict.Issues)
	}
	if refinementCalls != 1 {
		t.Errorf("expected 1 refinement call, got %d", refinementCalls)
	}
	if !strings.Contains(finalCode, "$1") {
		t.Errorf("expected revised code with parameterized query")
	}

	// 2. Unresolved issues after max iterations -> flags advisory notice without hanging
	staleCode := `exec.Command("sh", "-c", "rm -rf /")`
	finalFlaggedCode, verdictFlagged := critic.ExecuteCriticLoop(ctx, staleCode, func(ctx context.Context, code string, feedback string) (string, error) {
		return staleCode, nil // continues returning bad code
	})

	if verdictFlagged.IsClean {
		t.Errorf("expected flagged verdict")
	}
	if verdictFlagged.Iterations != 2 {
		t.Errorf("expected exactly 2 iterations clamped, got %d", verdictFlagged.Iterations)
	}
	if !strings.Contains(finalFlaggedCode, "TokMan Security Critic (`pool/security-tester`) Advisory") {
		t.Errorf("expected security advisory note in output: %s", finalFlaggedCode)
	}
}
