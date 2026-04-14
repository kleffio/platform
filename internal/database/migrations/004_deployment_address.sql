-- Add server_name and address fields to deployments.
-- server_name: the human-readable name used as container/pod name.
-- address:     host:port reported back by the daemon after provisioning.

ALTER TABLE deployments ADD COLUMN IF NOT EXISTS server_name TEXT NOT NULL DEFAULT '';
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS address     TEXT NOT NULL DEFAULT '';
