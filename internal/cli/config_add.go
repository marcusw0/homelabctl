package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/config"
)

type ConfigAddCmd struct {
	ConfigPath  string
	ServiceName string
}

func (c *ConfigAddCmd) Validate() error {
	if c.ConfigPath == "" {
		return errors.New("config path cannot be empty")
	}
	if c.ServiceName == "" {
		return errors.New("specify a name for the new service")
	}
	return nil
}

func (c *ConfigAddCmd) Run(ctx context.Context, streams IOStreams) error {
	newService, err := promptService(streams.In, streams.ErrOut, c.ServiceName)
	if err != nil {
		return err
	}

	newServiceV2 := fillDefaults(newService)

	if err := config.AddService(
		c.ConfigPath,
		c.ServiceName,
		newServiceV2,
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		streams.Out,
		"%s added to config\n",
		c.ServiceName,
	); err != nil {
		return err
	}
	return nil
}

func promptService(
	in io.Reader,
	out io.Writer,
	name string,
) (config.Service, error) {
	newService := config.Service{
		Enabled: true,
	}
	if _, err := fmt.Fprintf(
		out,
		"Enter the fqdn for %s: ",
		name,
	); err != nil {
		return newService,
			fmt.Errorf(
				"write fqdn: %w",
				err,
			)
	}
	if _, err := fmt.Fscan(in, &newService.FQDN); err != nil {
		return newService, fmt.Errorf("read FQDN: %w", err)
	}

	if err := check.ValidateHostname(newService.FQDN); err != nil {
		return newService, err
	}

	if _, err := fmt.Fprintf(
		out,
		"Enter %s's IP: ",
		name,
	); err != nil {
		return newService,
			fmt.Errorf(
				"write ip: %w",
				err,
			)
	}
	if _, err := fmt.Fscan(in, &newService.IP); err != nil {
		return newService, fmt.Errorf("read ip address: %w", err)
	}

	if err := check.ValidateIP(newService.IP); err != nil {
		return newService, err
	}

	if _, err := fmt.Fprintf(
		out,
		"Enter a port number for %s: ",
		name,
	); err != nil {
		return newService,
			fmt.Errorf(
				"write port: %w",
				err,
			)
	}
	if _, err := fmt.Fscan(in, &newService.Port); err != nil {
		return newService, fmt.Errorf("read port: %w", err)
	}

	if err := check.ValidatePort(newService.Port); err != nil {
		return newService, err
	}

	fmt.Fprint(out, "check your config file to see the new V2 additions\n")

	return newService, nil
}

func fillDefaults(service config.Service) config.Service {
	tls := service.EffectiveTLSWarnBefore()
	redirects := service.EffectiveFollowRedirects()
	return config.Service{
		Enabled:         service.Enabled,
		FQDN:            service.FQDN,
		IP:              service.IP,
		Port:            service.Port,
		Runbook:         service.Runbook,
		Checks:          service.EffectiveChecks(),
		Timeout:         service.EffectiveTimeout(),
		ExpectedStatus:  service.EffectiveStatusCode(),
		TLSWarnBefore:   &tls,
		FollowRedirects: &redirects,
		HTTPURL:         service.EffectiveHTTPURL(),
	}
}
