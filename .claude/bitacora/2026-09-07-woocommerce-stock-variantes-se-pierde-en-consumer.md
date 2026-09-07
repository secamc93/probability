**Ticket:** pendiente de crear en este momento | **Negocio:** Viga (reporto el sintoma), afecta a TODOS los negocios con productos variables en WooCommerce | **Canal:** WooCommerce

## Resumen

El push automatico de stock Probability -> WooCommerce nunca actualizo el stock de
NINGUNA variacion de producto (color/talla), desde que se implemento (2026-07-02).
El producto padre no controla el stock visible en un producto variable, asi que la
variacion se queda siempre con el estado por defecto de WooCommerce ("Hay
existencias"), sin importar cuanto cambie el stock real en Probability/Siigo.

## Sintoma reportado

Viga uso el plugin de WordPress "WebToffee Import Export (Pro)" para actualizar en
bloque variaciones de producto (color + talla, ej. `#14728 Lila 2XL`). Despues de
esa importacion, una variacion mostraba "Hay existencias" en vez de la cantidad real
que manda Probability. El cliente sospecho que el import le habia cambiado el ID del
producto en WooCommerce.

## Diagnostico

El import de WebToffee **no es la causa**, fue coincidencia de timing. El bug es de
Probability y es anterior al import de cualquier cliente:

1. El publisher SI arma el mensaje completo con el dato de la variante:
   `EcommerceStockPushMessage` (`back/central/services/modules/inventory/internal/
   domain/ports/ports.go:191-200`) incluye `ExternalVariantID string
   \`json:"external_variant_id,omitempty"\``, poblado desde
   `product_business_integrations.external_variant_id` (ver `adjust_stock.go`,
   `product_integration_queries.go`).

2. El consumer del lado WooCommerce (`.../woocommerce/internal/infra/primary/
   queue/inventory_push_consumer.go`) define su **propia** struct para deserializar
   el mismo mensaje JSON, y esa struct **no declaraba** el campo
   `external_variant_id`. `encoding/json` en Go descarta en silencio cualquier campo
   del JSON que no exista en la struct destino: sin error, sin log, sin rastro.

3. Resultado: al consumer solo le llegaba `ExternalProductID` (el ID del producto
   PADRE). `UpdateInventory` -> `UpdateProductStock` recibe ese ID crudo. El cliente
   HTTP (`update_product_stock.go:19`, funcion `splitVariationRef`) SI sabe construir
   el endpoint correcto de variacion (`PUT /products/{parent}/variations/{id}`)
   cuando recibe el formato compuesto `"parent:variacion"` -- pero nunca lo recibe,
   porque el dato se perdio un paso antes. El `PUT` real siempre cayo sobre
   `/products/{parent}`, que en un producto variable no tiene stock visible al
   comprador.

Esto ya estaba anotado como riesgo abierto en el diseno original: `.claude/alerts/
woocommerce-inventory-push.md:198` listaba "Manejo de variaciones en Woo (productos
variables) y su SKU por variacion" como decision pendiente de la Fase 1, nunca
cerrada.

### Segunda causa CONFIRMADA: mapeo de ID obsoleto (no era solo hipotesis)

Comparando en vivo el SKU real que reporto el cliente (`SD313-4XL`, "Sudadera Dama
Confianza y Estilo - Azul Oscuro - 4XL") vía `POST /woocommerce/inventory/compare`:

```
external_item_id: "10859:10863"
channel_qty: null
reason: "no se encontro la publicacion en el canal"
```

El ID que Probability tenia guardado para esa variacion (`10863`) ya no existe en
WooCommerce -- la publicacion fue eliminada/recreada, casi seguro por el import de
WebToffee (coincide con el reporte del cliente). **Se corrieron los dos bugs a la
vez** en este SKU: aunque el ID hubiera seguido siendo valido, la variante nunca
se habria actualizado por el bug del consumer; y aunque el consumer hubiera estado
bien, el ID guardado ya no apuntaba a nada.

Se corrio la comparacion completa del catalogo de Viga (2254 SKUs, 23 paginas):
**617 (27%) tenian el mismo problema** ("no se encontro la publicacion en el
canal") -- un drift grande y sostenido en el tiempo, no un incidente puntual.
Se corrigieron los 617 en lote via `POST /woocommerce/products/associate`
(re-matchea por SKU contra el catalogo real de WooCommerce y corrige el ID
guardado) + `POST /woocommerce/inventory/sync` para el SKU de prueba, confirmando
que `SD313-4XL` quedo en `channel_qty: 11 = probability_qty: 11, delta: 0`.

### Hipotesis descartadas

- Ninguna: las dos causas sospechadas resultaron reales y compuestas para este caso.

## Correccion (dos partes)

**1. Variante perdida en el consumer.** `ecommerceStockPushMessage` (consumer)
ahora declara `ExternalVariantID` igual que el mensaje del publisher, y `handle()`
compone `"parent:variacion"` antes de llamar a `UpdateInventory` cuando el mensaje
trae variante.
Archivo: `back/central/services/integrations/ecommerce/woocommerce/internal/infra/
primary/queue/inventory_push_consumer.go`

**2. Mapeo de ID obsoleto -- auto-correccion permanente.** Se evaluaron dos
disenos: (a) resolver por SKU en CADA push, o (b) push directo por ID guardado y
auto-corregir solo cuando falla. Se eligio (b) por costo: (a) le suma 2-3 llamadas
a la API de WooCommerce a CADA cambio de stock (un `GET` de busqueda + navegar
variaciones antes del `PUT`), mientras que (b) mantiene el caso normal en 1 sola
llamada y solo paga el costo extra cuando de verdad hay un ID roto. Coincide con
lo que ya estaba anotado como plan pendiente en `.claude/alerts/
woocommerce-inventory-push.md` ("Fase 2: resolucion por SKU, si hace falta" --
este caso de Viga es el "si hace falta").

Flujo implementado en `update_inventory.go`:
1. `UpdateProductStock` devuelve `domain.ErrProductNotFoundInStore` (404) en vez de
   un error de texto plano (antes no era detectable por tipo).
2. `healStaleMapping`: busca en `product_business_integrations` el producto de
   Probability que tenia ese ID guardado (`GetProductIDByExternalRef`, nuevo),
   obtiene su SKU (`GetProductSKUByID`, nuevo), y llama a `AssociateProducts` (ya
   existia, es el mismo mecanismo del boton "Comparar productos") **solo para ese
   SKU** -- corrige el ID guardado re-matcheando contra el catalogo real de
   WooCommerce.
3. Si la re-asociacion encontro un ID nuevo (`GetExternalRefs`, nuevo), reintenta
   el `PUT` **una sola vez** con el ID corregido. Si no hay ID nuevo o el reintento
   tambien falla, se loguea como antes (sin loop).

Nuevos metodos en `IProductRepository` (mismo modulo, sin tocar otros): 
`GetProductSKUByID`, `GetProductIDByExternalRef`, `GetExternalRefs`.
Nuevo error tipado: `domain.ErrProductNotFoundInStore`.

## Verificacion

`go build ./...` limpio en todo el backend. Tests unitarios nuevos, todos pasan:

- `inventory_push_consumer_test.go` (2 casos): con variante compone `"123:456"`;
  sin variante no rompe el caso simple.
- `update_inventory_test.go` (2 casos): sin mapeo encontrado devuelve el error
  original (no enmascara el fallo real); con mapeo encontrado, ejecuta el flujo
  completo de re-asociacion + reintento (usa el mecanismo real de
  `AssociateProducts`/`productmatch.Reconcile`, no un stub) y confirma 2 llamadas
  a `UpdateProductStock` (original + reintento con el ID corregido).

**Verificado tambien en produccion, en vivo, contra la tienda real de Viga**
(no solo en test): el SKU `SD313-4XL` tenia el ID obsoleto (`10859:10863`,
"no se encontro la publicacion en el canal"); se corrio `AssociateProducts` +
`SyncInventory` manualmente y quedo en `channel_qty: 11 = probability_qty: 11,
delta: 0, unchanged`. El fix de codigo automatiza este mismo mecanismo para que
no dependa de una intervencion manual la proxima vez.

## Impacto

Afecta a **todo negocio con productos variables (color/talla/etc.) en WooCommerce**,
no solo a Viga -- el bug del consumer esta desde julio (Fase 1 del push). El bug
del mapeo obsoleto es mas amplio: cualquier negocio que use una herramienta externa
de import/export en su tienda (WebToffee u otra) puede generarlo, sin relacion con
variantes.

**Medido en Viga (business 46): 617 de 2254 SKUs/variantes (27%) tenian el ID
obsoleto**, corregidos en lote el 2026-09-07. No se corrio el mismo diagnostico en
otros negocios con WooCommerce -- pendiente.

## Pendientes

- Ticket: **TKT-000076**, creado.
- Correr el mismo diagnostico (`POST /woocommerce/inventory/compare` sin filtro de
  SKU, buscar `reason: "no se encontro la publicacion en el canal"`) en los demas
  negocios con integracion WooCommerce activa, para saber si el drift de Viga es
  un caso aislado o generalizado.
- Verificar E2E contra una tienda WooCommerce de pruebas (no solo produccion real)
  despues del deploy, con un producto variable nuevo.
- Revisar si aplica el mismo patron de "struct propia que descarta campos nuevos
  silenciosamente" en otros consumers de colas del proyecto (Shopify, MercadoLibre,
  Jumpseller, TiendaNube tienen su propio `UpdateInventory`/consumer de push de
  stock) -- es un riesgo estructural: cualquier campo nuevo agregado al mensaje del
  publisher requiere acordarse de replicarlo en CADA consumer que lo lea.
- El auto-heal implementado corrige de a un SKU por vez (el que fallo). Si un
  negocio tiene un drift grande como el de Viga (27%), cada 404 dispara su propia
  re-asociacion (cada una relee el catalogo COMPLETO de WooCommerce). Aceptable
  como salvavidas para fallos aislados; si hay una rafaga grande de fallos
  simultaneos (ej. una migracion masiva del cliente), vale la pena un job de
  reconciliacion periodica en vez de depender solo del self-heal reactivo -- ya
  estaba anotado como pendiente en la alerta original.
