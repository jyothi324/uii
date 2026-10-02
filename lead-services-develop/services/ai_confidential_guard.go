package services

import "regexp"

// confidentialIntentPattern is a coarse, keyword-based check for questions
// that are asking for credentials or other confidential account data
// (passwords, secrets, tokens, TOTP/MFA codes, etc.), run against the raw
// natural-language question BEFORE it ever reaches Gemini.
//
// This matters independently of the SQL-level guardrail
// (validateAgentsColumnAccess in sql_guardrail.go): AISchemaContext never
// mentions the password/password_hashed/details columns, so Gemini
// typically won't generate SQL that references them at all for a question
// like "what is the password for agent X" — it just queries what it does
// know about, finds nothing relevant, and phrases a technically-true but
// misleading answer like "there is no password on record for that agent."
// That's worse than an honest refusal: it implies no such data exists
// rather than that Bloom won't share it. This check catches the intent
// before generation so the user gets a direct, honest decline instead.
var confidentialIntentPattern = regexp.MustCompile(
	`(?i)\b(password|passwd|pwd|secret|totp|mfa\s*code|2fa\s*code|api[\s_-]?key|` +
		`access[\s_-]?token|auth[\s_-]?token|credential|ssn|social security|` +
		`credit card|card number)\b`,
)

// IsConfidentialDataRequest reports whether a user's natural-language
// question is asking for the kind of confidential data Bloom must never
// surface, independent of what SQL might eventually get generated for it.
func IsConfidentialDataRequest(question string) bool {
	return confidentialIntentPattern.MatchString(question)
}
