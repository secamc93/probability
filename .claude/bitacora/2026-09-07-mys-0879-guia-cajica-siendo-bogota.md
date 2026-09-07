**Ticket:** pendiente de crear (bloqueado el acceso a produccion para crearlo via API en esta sesion; ver pendientes) | **Negocio:** Mystic Rose - Official (business 52 en local) | **Canal:** panel, orden manual

## Resumen

La orden MYS-0879 se creo con "Ciudad y Departamento" = BOGOTA (CUNDINAMARCA),
pero la guia (ENVIA) salio con destino CAJICA - CUNDINAMARCA. Causa: la
resolucion de geozona de la orden confiaba ciegamente en las coordenadas
geocodificadas, que Google devolvio dentro del municipio vecino.

## Sintoma

Modal "Editar Orden" de MYS-0879: direccion "calle 53sur #18c48", "Ciudad y
Departamento" = BOGOTA (CUNDINAMARCA). La guia impresa de ENVIA para la misma
orden trae "CAJICA - CUNDINAMARCA" como destino, con la misma direccion y el
mismo destinatario (Maria Neila Alonso).

## Diagnostico

Cadena de datos confirmada por lectura de codigo (sin acceso a la BD de
produccion en esta sesion, ver Pendientes):

1. `usecasecreateorder/geocode_order.go:geocodeOrderIfNeeded` geocodifica la
   direccion con Google (`geocoder.go:38-40`, solo `components=country:co`,
   sin acotar por ciudad/departamento) y guarda `shipping_lat/lng`.
2. `ResolveOrderGeozone` (`orders/internal/infra/secondary/repository/
   geozone_queries.go`) resolvia la geozona **solo por coordenadas**
   (`ST_Contains` sobre `shipping_lat/lng`), y solo caia al texto
   (`shipping_city`/`shipping_state`) si el punto no matcheaba nada.
3. `repository.go:109` (`getGeozoneCode`) convierte `geozone_city_id` en
   `DestinationDaneCode`, que el frontend usa como fuente de verdad
   ("Nivel 1" en `dane-lookup.ts`, prioriza `destination_dane_code` sobre el
   texto) para prellenar/validar el destino de la guia.

Como el punto que devolvio Google si cayo dentro del poligono de Cajica, el
fallback por texto (que habria acertado Bogota) nunca corrio, y el codigo
DANE de Cajica quedo como fuente de verdad aunque el texto en pantalla
siempre dijo Bogota.

Mismo mecanismo que `2026-09-02-orden-geocodificada-en-buga-siendo-zarzal.md`
(coordenada gana sobre texto sin cross-check), con disparador distinto: alla
las coordenadas quedaron pegadas de una sugerencia vieja; aca el geocode se
ejecuto de nuevo correctamente pero Google devolvio un punto en el municipio
vecino. El fix de esa fecha (`handleCitySelect` re-geocodifica al corregir la
ciudad) sigue funcionando, pero no cubre este caso: el re-geocode se
ejecuto, solo que el resultado externo fue malo y nada lo valido contra el
texto ya confirmado.

### Hipotesis descartadas

- **Bug del carrier o de EnvioClick**: no, ENVIA imprimio exactamente el
  `daneCode` que se le mando; el dato malo se origino en nuestro sistema.
- **El fix del 02/09 no se aplico**: si se aplico y sigue funcionando (ver
  codigo), simplemente no alcanza para este caso porque el problema no es
  "falta re-geocodificar", es "el geocode devolvio un punto malo y nadie lo
  valido contra el texto".

## Correccion

`ResolveOrderGeozone` (orders) y `ResolveShipmentGeozone` (shipments,
duplicado por aislamiento de modulos) ahora calculan **dos** resoluciones
independientes -punto (ST_Contains) y texto (shipping_city/shipping_state)- y
las comparan a nivel ciudad:

- Si ambas existen y **discrepan**, gana el texto (es lo que el usuario vio y
  confirmo). Se descartan el punto (`destination_point = NULL`) y los niveles
  finos (barrio/localidad/admin_district) derivados de el, porque quedaron
  atados a una ubicacion que no es la real.
- Si coinciden, o el texto no resuelve nada (ciudad no catalogada), se
  mantiene el comportamiento anterior (punto con fallback a texto).

Archivos:
- `back/central/services/modules/orders/internal/infra/secondary/repository/geozone_queries.go`
- `back/central/services/modules/shipments/internal/infra/secondary/repository/geozone_queries.go`

No se toco el geocoder externo (Google) ni la logica de checkout Woo/Shopify
(`2026-09-03-ciudad-dane-comodin-prefijo.md`), que no aplica aqui: esta orden
es manual, no de un canal.

## Verificacion

Sin acceso a produccion en esta sesion (bloqueos del modo automatico, ver
Pendientes). Se compilo (`go build ./...`, limpio) y se corrieron 5 casos
simulados contra la copia LOCAL (business 52, geozonas sinteticas Bogota/
Cajica con poligonos propios, ejecutando `ResolveOrderGeozone` real vía un
test temporal desechado despues, no forma parte del repo):

| Caso | Texto | Punto | Resultado esperado | Resultado obtenido |
|---|---|---|---|---|
| T1 (repro exacto de MYS-0879) | BOGOTA | dentro de Cajica | Bogota | **Bogota (11001)** -- antes habria dado Cajica |
| T2 | BOGOTA | dentro de Bogota | Bogota | Bogota (11001) |
| T3 | CAJICA | dentro de Cajica | Cajica | Cajica (25126) -- no rompe pedidos legitimos a Cajica |
| T4 | BOGOTA | sin coordenadas | Bogota | Bogota (11001) -- fallback por texto preservado |
| T5 | ciudad no catalogada (SUBACHOQUE) | dentro de Cajica | Cajica (no hay con que comparar) | Cajica (25126) |

Los 5 casos pasaron. Datos de prueba limpiados al terminar (0 filas
residuales en `orders`/`geozones` de test).

## Pendientes

- **Ticket**: TKT-000074, creado y en `testing`.
- **Verificado en LOCAL** (2026-09-07): el usuario creo una orden nueva con
  los mismos datos de MYS-0879 (misma direccion, ciudad BOGOTA CUNDINAMARCA,
  negocio Mystic Rose) y confirmo que la geozona/destino ahora resuelve
  Bogota. Falta la verificacion en produccion una vez desplegado el commit
  `cbeb0fe6` (mergeado en `1f41ed70`, pusheado a `main`) para pasar el ticket
  a `resolved`.
- **Dato historico de MYS-0879**: la guia ya salio con Cajica; no se corrigio
  (se coordina aparte con el cliente si hace falta reenvio).
- Considerar si vale la pena, ademas, acotar el geocoder de Google con
  `bounds`/`components=locality:` cuando ya se conoce la ciudad elegida, para
  reducir la tasa de puntos fuera de la ciudad esperada (mitigacion en la
  fuente, complementaria al cross-check ya aplicado).
