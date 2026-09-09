package utils

import "strings"

func GenerateSlug(input string) string {
	// Implement your slug generation logic here
	// For example, you can convert the input to lowercase, replace spaces with hyphens, and remove special characters
	// This is a simple example; you may want to use a more robust library for slug generation in production
	slug := strings.ToLower(input)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	return slug
}
