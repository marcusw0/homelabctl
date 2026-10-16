# Grafana runbook

Use this checklist when the dashboard reports an unhealthy service.

## First checks

1. Open the service details in `homelabctl dashboard`.
2. Read the HTTP and TCP results.
3. Check whether other services on the same host are affected.

## Connection refused

The host responded, but no process accepted the connection.

- Confirm the service is running.
- Check that its listening port matches the inventory.
- Review the reverse proxy upstream address.

## Unexpected HTTP status

A reachable HTTP endpoint can still be unhealthy.

- Confirm the configured health endpoint.
- Check authentication and redirect settings.
- Compare the response with the expected status code.

## Recovery checklist

1. Identify the failure before changing settings.
2. Restore the service or correct its configuration.
3. Run a fresh service check.
4. Confirm the dashboard returns to healthy.

## Follow-up

Record what happened so the next investigation is easier.

- Symptom: service health check failed.
- Cause: investigate using the checklist above.
- Resolution: pending.
