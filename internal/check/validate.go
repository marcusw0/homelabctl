package check

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
)

func ValidateAddrPort(combined string) error {
	host, port, err := net.SplitHostPort(combined)
	if err != nil {
		return err
	}
	if host == "" {
		return errors.New("host cannot be empty")
	}
	uintPort, err := strconv.ParseUint(port, 10, 16)
	if err != nil || uintPort == 0 {
		return errors.New("port number must be any number from 1-65535")
	}

	return ValidateIPOrHostname(host)
}

func ValidatePort(port int) error {
	if port <= 0 || port > 65535 {
		return errors.New("port number must be any number from 1-65535")
	}

	return nil
}

func ValidateIP(ip string) error {
	_, err := netip.ParseAddr(ip)
	if err != nil {
		return err
	}

	return nil
}

func ValidateHostname(host string) error {
	host = strings.TrimSuffix(host, ".")

	if host == "" || len(host) > 253 {
		return fmt.Errorf("%q is not a valid hostname", host)
	}

	labels := strings.Split(host, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return fmt.Errorf("%q is not a valid hostname", host)
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("%q is not a valid hostname", host)
		}

		for _, char := range label {
			valid := char >= 'a' && char <= 'z' ||
				char >= 'A' && char <= 'Z' ||
				char >= '0' && char <= '9' ||
				char == '-'
			if !valid {
				return fmt.Errorf("%q is not a valid hostname", host)
			}
		}
	}

	lastLabel := labels[len(labels)-1]
	for _, char := range lastLabel {
		if char >= 'a' && char <= 'z' ||
			char >= 'A' && char <= 'Z' {
			return nil
		}
	}

	return fmt.Errorf("%q is not a valid hostname", host)

}

func ValidateIPOrHostname(host string) error {
	if _, err := netip.ParseAddr(host); err == nil {
		return nil
	}

	return ValidateHostname(host)
}

func NormalizeHTTPURL(target string) (string, error) {
	if index := strings.Index(target, "://"); index >= 0 {
		scheme := target[:index]
		if !strings.EqualFold(scheme, "http") &&
			!strings.EqualFold(scheme, "https") {
			return "", fmt.Errorf("unsupported scheme %q", scheme)
		}
		target = strings.ToLower(scheme) + target[index:]
	} else {
		target = "https://" + target
	}

	parsed, err := url.ParseRequestURI(target)
	if err != nil {
		return "", fmt.Errorf("invalid HTTP target %q: %w", target, err)
	}
	if err := ValidateIPOrHostname(parsed.Hostname()); err != nil {
		return "", fmt.Errorf("invalid HTTP target %q: %w", target, err)
	}
	strPort := parsed.Port()
	if strPort != "" {
		port, err := strconv.Atoi(strPort)
		if err != nil {
			return "", err
		}
		err = ValidatePort(port)
		if err != nil {
			return "", err
		}
	}

	return target, nil
}
