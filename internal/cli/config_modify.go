package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/marcusw0/homelabctl/internal/config"
	"github.com/marcusw0/homelabctl/internal/tui/serviceform"
)

func runConfigAdd(ctx context.Context, request ModifyConfigRequest, streams IOStreams) error {
	if request.ConfigPath == "" {
		return errors.New("config path cannot be empty")
	}
	if request.ServiceName == "" {
		return errors.New("specify a name for the new service")
	}

	newService, err := serviceform.Run(
		ctx,
		streams.In,
		streams.ErrOut,
		request.ServiceName,
		nil,
	)
	if err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	newServiceV2 := fillDefaults(newService)

	if err := config.AddService(
		request.ConfigPath,
		request.ServiceName,
		newServiceV2,
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		streams.Out,
		"%s added to config\n",
		request.ServiceName,
	); err != nil {
		return err
	}
	return nil
}

func runConfigEdit(ctx context.Context, request ModifyConfigRequest, streams IOStreams) error {
	if request.ConfigPath == "" {
		return errors.New("config path cannot be empty")
	}
	if request.ServiceName == "" {
		return errors.New("specify a name for the new service")
	}
	cfg, err := config.Load(request.ConfigPath)
	if err != nil {
		return err
	}

	existing, ok := cfg.Services[request.ServiceName]
	if !ok {
		return fmt.Errorf("service %q is not in config", request.ServiceName)
	}

	newService, err := serviceform.Run(
		ctx,
		streams.In,
		streams.ErrOut,
		request.ServiceName,
		&existing,
	)
	if err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	newServiceV2 := fillDefaults(newService)

	if err := config.UpdateService(
		request.ConfigPath,
		request.ServiceName,
		newServiceV2,
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		streams.Out,
		"changes to %q have been saved",
		request.ServiceName,
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
