# terraform-provider-gestioip
Terrform provider for Marc Uebel's GestioIP

## API notes

The official GestioIP 3.5 API guide documents `.../gestioip/api/api.cgi` with Basic Auth.

When testing against the clean `gestioip/gestioip:3570` container image on `http://localhost` on March 18, 2026, the observed behavior was different:

- `/gestioip/api/api.cgi` was not exposed
- `/gestioip/intapi.cgi` was exposed
- `/gestioip/intapi.cgi` required a session cookie flow instead of working with Basic Auth alone
- the observed internal API surface was narrower, and `listNetworks` was confirmed to work for network reads

The provider currently contains endpoint fallback logic so it can work against both the documented API path and the container image behavior observed above.

## Current network approach

For the free container image tested on March 18, 2026, the provider uses a hybrid integration model for networks:

- reads use `intapi.cgi` and `listNetworks`
- writes use the frontend CGI flow:
  - `res/ip_insertred.cgi`
  - `res/ip_modred.cgi`
  - `res/ip_deletered.cgi`

For hosts in the same free container image, the provider currently uses the frontend CGI flow end to end:

- reads parse `ip_show.cgi`
- create and update use `res/ip_modip.cgi`
- delete uses `res/ip_deleteip.cgi`

This means `gestioip_network` is currently implemented against the real behavior of the free edition, not only against the paid API guide.
The same is now true for `gestioip_host`.

Two practical notes from the validated container setup:

- `client_name` must be resolvable to a GestioIP client visible in the UI
- `site` and `category` should match values that already exist in GestioIP, because the frontend flow expects those catalogs to be preconfigured

The repository also includes an env-gated integration test for the network lifecycle in `internal/client/integration_test.go`.
