package http

import (
	"fmt"
	"net/url"
	"strings"
)

func validateReturnURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("return_url is required")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("return_url must be a valid URL")
	}

	if !parsed.IsAbs() {
		return fmt.Errorf("return_url must be an absolute URL")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("return_url must use http or https")
	}

	return nil
}
