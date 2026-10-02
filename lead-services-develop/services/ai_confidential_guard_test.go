package services

import "testing"

func TestIsConfidentialDataRequest_MatchesCredentialIntent(t *testing.T) {
	cases := []string{
		"what is the password for an agent's shankar",
		"show me the passwd for shankar",
		"what's the TOTP code for admin",
		"give me the api key",
		"what is the access token",
		"what is the SSN on file for this customer",
	}
	for _, q := range cases {
		if !IsConfidentialDataRequest(q) {
			t.Errorf("expected %q to be flagged as a confidential data request", q)
		}
	}
}

func TestIsConfidentialDataRequest_AllowsOrdinaryQuestions(t *testing.T) {
	cases := []string{
		"how many leads are rejected?",
		"what is the price of the roofing job?",
		"who is assigned to this lead?",
		"show me all estimates from last month",
	}
	for _, q := range cases {
		if IsConfidentialDataRequest(q) {
			t.Errorf("expected %q not to be flagged as a confidential data request", q)
		}
	}
}
