package cli

import (
	"context"
	"fmt"
)

func Dispatch(
	ctx context.Context,
	req Request,
	streams IOStreams,
) error {
	switch req.Kind {
	case CmdHTTPCheck:
		return runHTTPCheck(ctx, req.HTTP, streams)
	case CmdDNSCheck:
		return runDNSCheck(ctx, req.Check, streams)
	case CmdServiceCheck:
		return runServiceCheck(ctx, req.Check, req.ConfigPath, streams)
	case CmdTCPCheck:
		return runTCPCheck(ctx, req.Check, streams)
	case CmdTLSCheck:
		return runTLSCheck(ctx, req.Check, streams)
	case CmdAllCheck:
		return runAllCheck(ctx, req.ConfigPath, streams)
	case CmdAddConfig:
		return runConfigAdd(ctx, req.Modify, streams)
	case CmdEditConfig:
		return runConfigEdit(ctx, req.Modify, streams)
	case CmdInitConfig:
		return runConfigInit(req.ConfigPath, streams)
	case CmdListConfig:
		return runList(req.ConfigPath, streams)
	case CmdDashboard:
		return runDashboard(ctx, req.ConfigPath)
	case CmdRunbook:
		return runRunbook(req.Runbook, streams)
	}
	return fmt.Errorf("unsupported command kind: %d", req.Kind)
}
