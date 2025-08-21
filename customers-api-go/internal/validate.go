package main

import "strings"

func ValidateCustomer(c Customer) bool {
	return strings.TrimSpace(c.Name) != "" && strings.Contains(c.Email, "@")
}
