package hw10programoptimization

import (
	"errors"
	"io"
	"strings"
)

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	ur := NewEmailReader(r)
	return countDomains(ur, domain)
}

func countDomains(r *EmailReader, domain string) (DomainStat, error) {
	// Add dot and lower-case domain to check the suffix of email later
	domain = "." + strings.ToLower(domain)

	result := make(DomainStat)
	for {
		// Try to read email
		email, err := r.NextEmail()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		// Check if the email domain match to the asked domain.
		if email != "" && strings.HasSuffix(email, domain) {
			idx := strings.Index(email, "@")
			if idx >= 0 {
				// Ignore '@'
				result[email[idx+1:]]++
			}
		}
	}

	return result, nil
}
