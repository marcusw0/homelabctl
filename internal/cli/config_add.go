package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/marcusw0/homelabctl/internal/config"
	"github.com/marcusw0/homelabctl/internal/tui/serviceform"
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
	newService, err := serviceform.Run(ctx, streams.In, streams.ErrOut, c.ServiceName)
	if err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
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
