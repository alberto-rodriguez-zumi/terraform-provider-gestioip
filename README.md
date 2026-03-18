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
