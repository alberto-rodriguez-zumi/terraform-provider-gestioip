# AGENTS.md

## Objetivo del repositorio

Este repositorio implementa un provider de Terraform para interactuar con la API de GestioIP.

La base técnica del provider debe ser:

- Go
- Terraform Plugin Framework
- un cliente HTTP interno pequeño y explícito

La meta no es envolver toda la API de una vez, sino construir un provider mantenible, predecible y alineado con la experiencia habitual de Terraform.

## Fuente funcional principal

La referencia funcional principal es la guía oficial de la API de GestioIP 3.5:

- PDF: <https://www.gestioip.net/docu/GestioIP_3.5_API_Guide.pdf>

Puntos importantes confirmados en esa guía:

- La API se expone en `.../gestioip/api/api.cgi`.
- La autenticación se realiza con Basic Auth.
- Las peticiones se envían como parámetros `attribute=value` en formato URL, normalmente mediante `POST`.
- El formato de salida por defecto es XML, por lo que el provider debe solicitar JSON siempre que sea posible usando `output_type=json`.
- Muchas operaciones requieren `client_name`.
- Existen operaciones para `hosts`, `networks`, `vlans`, `vlan providers`, `clients` y varias acciones de discovery.
- Para listados de redes existe el parámetro `no_csv=yes`, que devuelve una estructura más adecuada para el provider que la salida CSV embebida.

Si en algún punto el comportamiento observado en una instancia real difiere del PDF, documentarlo en el código o en el README y priorizar el comportamiento real verificado.

Observación importante ya verificada contra una instancia local limpia de la imagen `gestioip/gestioip:3570` el 18 de marzo de 2026:

- la ruta documentada `/gestioip/api/api.cgi` no estaba expuesta en esa imagen
- la ruta operativa observada fue `/gestioip/intapi.cgi`
- `intapi.cgi` requirió sesión por cookie y no funcionó con Basic Auth puro
- la superficie observada en `intapi.cgi` fue más reducida; para redes se confirmó `listNetworks`
- la imagen sí incluye CGI de frontend para escritura de redes como `res/ip_insertred.cgi`, `res/ip_modred.cgi` y `res/ip_deletered.cgi`, pero siguen un flujo de formularios web y no un contrato API estable equivalente a `intapi.cgi`

Esto significa que el provider debe tolerar al menos dos variantes de despliegue:

- la API documentada del PDF
- la API interna expuesta por la imagen de contenedor probada

Cuando se implemente o ajuste una entidad, priorizar siempre la comprobación contra una instancia real además del PDF.
Si en algún caso se decide soportar escritura vía CGI de frontend, tratarlo como una integración distinta y documentar muy bien sus límites y supuestos.

## Principios de diseño

- Usar Terraform Plugin Framework, no Plugin SDK v2.
- Priorizar recursos y data sources estables antes que cubrir endpoints muy procedimentales.
- Modelar en Terraform entidades declarativas; evitar exponer acciones efímeras como recursos si no representan estado duradero.
- Preferir identificadores naturales estables cuando existan, pero conservar el ID interno de GestioIP si simplifica lecturas, importaciones o actualizaciones.
- Hacer `Read` después de `Create` y `Update` para normalizar el estado.
- No ocultar rarezas de la API: si un endpoint es ambiguo, reflejar esa restricción en schema, validaciones o documentación.

## Prioridades funcionales

Orden recomendado de implementación:

1. Provider configuration y cliente API.
2. Data source o recurso básico de `network`.
3. Recursos y data sources de `host`.
4. Recursos y data sources de `vlan`.
5. Recursos y data sources de `vlan provider`.
6. Recursos y data sources de `client`.
7. Acciones especiales como first free IP/network, discovery o reservas automáticas, solo cuando el modelo Terraform esté claro.

## Convenciones para la integración con la API

- Centralizar el acceso HTTP en `internal/client`.
- Toda petición debe declarar explícitamente:
  - `request_type`
  - `output_type=json` cuando el endpoint lo soporte
  - `client_name` cuando aplique
- Preferir `POST` con `application/x-www-form-urlencoded`.
- Tratar respuestas con campo `error` como fallo funcional aunque el HTTP status sea `200`.
- Añadir logs con `tflog` solo para contexto útil y sin exponer contraseñas ni datos sensibles.
- Normalizar `base_url` para evitar dobles barras y construir internamente la ruta final a `api.cgi`.
- Permitir detección o fallback de endpoint cuando el despliegue use `intapi.cgi` en lugar de `api/api.cgi`.
- Si el endpoint real requiere cookie de sesión, encapsular ese flujo dentro del cliente y no en recursos o data sources.

## Diseño del provider

El provider debe empezar con una configuración pequeña y clara. La configuración inicial esperada es:

- `base_url`
- `username`
- `password`
- `client_name` como candidato fuerte a valor de provider por defecto

`client_name` merece atención especial:

- La guía de la API muestra que la mayoría de operaciones lo exigen.
- `createClient` parece ser la excepción más visible.
- Si se define a nivel de provider, los recursos y data sources pueden permitir override opcional cuando tenga sentido.

Antes de implementar muchos recursos, conviene fijar esta decisión y aplicarla de forma consistente.

## Diseño de recursos y data sources

- Los nombres deben seguir el patrón `gestioip_<entity>`.
- Los data sources deben cubrir primero búsquedas directas y deterministas.
- Evitar data sources basados en resultados ambiguos, por ejemplo búsquedas por hostname no único, salvo que el schema haga explícita esa limitación.
- Los recursos deben mapear con precisión los campos soportados por la API y no inventar atributos.
- Si una operación existe en el PDF pero no está disponible en la variante real que estamos probando, documentarlo y no asumir soporte inmediato.
- Los campos calculados deben usarse solo cuando realmente provengan del servidor o sean necesarios para estabilidad del estado.
- Los campos sensibles deben marcarse como `Sensitive`.
- Si la API usa nombres poco idiomáticos como `new_BM`, eso debe quedar encapsulado en el cliente o en la capa de mapeo, no en el schema público de Terraform.

## Manejo de identificadores

Siempre que sea viable, usar un ID de estado canónico y estable.

Opciones aceptables según entidad:

- ID interno de GestioIP
- clave natural única, por ejemplo `client_name/ip`
- una composición documentada si la API obliga a leer por varios campos

El formato elegido debe ser consistente dentro de cada recurso y quedar reflejado en la lógica de importación.

## Custom columns

La API soporta custom columns en varias entidades.

Reglas para esta primera etapa:

- No introducir complejidad dinámica innecesaria en el schema hasta verificar bien el comportamiento real.
- Si se exponen, preferir `Map` de strings o estructuras bien delimitadas.
- Documentar claramente cualquier limitación derivada de nombres libres o columnas no tipadas.
- Tener especial cuidado con casos especiales documentados, como `new_vlan_id` al crear redes.

## Acciones especiales y endpoints procedimentales

Hay endpoints como:

- primer prefijo libre dentro de una root network
- primera IP libre
- reserva de primera IP libre
- discovery vía SNMP o DNS

Estos endpoints no encajan siempre bien como recursos declarativos.

Reglas:

- Evaluar primero si deben ser data sources.
- Solo convertirlos en recursos si representan estado persistente y reconciliable.
- Si un endpoint produce resultados inherentemente cambiantes, documentar el riesgo de drift o evitar modelarlo como recurso.

## Estructura del código

Estructura objetivo:

- `main.go` para arranque del provider
- `internal/provider` para provider, recursos y data sources
- `internal/client` para transporte HTTP, requests, responses y helpers de error
- `examples/` para configuraciones de uso

Cuando el número de entidades crezca:

- separar un archivo por recurso
- separar un archivo por data source
- mantener tipos de request/response cerca del cliente o del dominio, pero no duplicados sin necesidad

## Testing

Mínimos esperados:

- tests unitarios para parseo de respuestas y manejo de errores de API
- tests unitarios para expand/flatten entre schema y payloads
- tests básicos del provider

Cuando la base esté madura:

- tests de aceptación opcionales contra una instancia real o de laboratorio de GestioIP
- los acceptance tests deben quedar detrás de variables de entorno y no ejecutarse por defecto

No asumir que una respuesta HTTP 200 implica éxito de negocio; esto debe tener cobertura de tests.
No asumir tampoco que todas las instalaciones de GestioIP exponen exactamente el mismo endpoint o el mismo conjunto de `request_type`.

## Estilo de implementación

- Mantener funciones pequeñas y con nombres claros.
- Evitar abstracciones genéricas prematuras.
- Preferir tipos explícitos de request y response cuando mejoren legibilidad.
- Encapsular las rarezas de la API en un sitio, idealmente el cliente o helpers de conversión.
- Añadir comentarios solo cuando expliquen una decisión no obvia.

## Documentación

Cada nuevo recurso o data source debería ir acompañado de:

- documentación de uso en `README.md` o en `docs/` si más adelante añadimos generación de docs
- ejemplo mínimo en `examples/` cuando aporte valor
- notas sobre limitaciones conocidas de la API si afectan al comportamiento Terraform

## Flujo de contribución

- Mantener cambios pequeños y enfocados.
- Ejecutar `gofmt -w .` y `go test ./...` antes de cerrar una tarea, usando cachés locales si el entorno lo requiere.
- No mezclar refactors amplios con nuevas entidades de la API en el mismo cambio si se puede evitar.
- Si una decisión de modelado no es evidente, documentarla aquí o en el README antes de escalar la implementación.

## Decisiones abiertas

Estas decisiones deben revisarse temprano antes de ampliar mucho la superficie del provider:

- si `client_name` será obligatorio en el provider, opcional con override por recurso, o obligatorio por recurso
- qué recursos iniciales tendrán soporte de `ImportState`
- cómo exponer `customColumns`
- si conviene añadir `docs/` generada más adelante
- qué endpoints especiales merecen data source y cuáles no deben entrar en la primera versión
- cómo separar de forma limpia las capacidades comunes frente a las específicas de `intapi.cgi`
