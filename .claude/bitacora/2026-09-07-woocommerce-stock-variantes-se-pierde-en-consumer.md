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

### Hipotesis descartadas

- **El import de WebToffee cambio el ID del producto en WooCommerce.** No se pudo
  confirmar ni descartar con datos de produccion (mismo bloqueo de acceso de la
  sesion), pero es irrelevante para el sintoma: aunque el ID nunca hubiera cambiado,
  la variacion JAMAS habia recibido push de stock, porque el campo se perdia en el
  consumer antes de llegar a la API de WooCommerce.

## Correccion

`ecommerceStockPushMessage` (consumer) ahora declara `ExternalVariantID` igual que
el mensaje del publisher, y `handle()` compone `"parent:variacion"` antes de llamar
a `UpdateInventory` cuando el mensaje trae variante.

Archivo: `back/central/services/integrations/ecommerce/woocommerce/internal/infra/
primary/queue/inventory_push_consumer.go`

## Verificacion

`go build ./...` limpio. Se agrego test unitario nuevo
(`inventory_push_consumer_test.go`, no existia cobertura previa de este consumer):
un caso con `external_variant_id` (confirma que arma `"123:456"`) y uno sin el
(confirma que un producto simple sigue funcionando igual, sin variacion). Ambos
pasan.

No se pudo verificar E2E contra una tienda WooCommerce real con productos variables
en esta sesion (bloqueo de acceso a produccion).

## Impacto

Afecta a **todo negocio con productos variables (color/talla/etc.) en WooCommerce**,
no solo a Viga -- el bug esta desde julio (implementacion de la Fase 1 del push).
No se pudo medir cuantos negocios/SKUs tienen variantes activas en Woo en esta
sesion (mismo bloqueo de acceso a produccion). Pendiente cuantificar cuando haya
acceso de lectura.

## Pendientes

- Crear el ticket formal (bloqueo de acceso a produccion via API en el momento de
  escribir esta entrada -- reintentar).
- Verificar E2E contra una tienda WooCommerce real con un producto variable
  despues del deploy.
- Cuantificar el impacto historico (cuantos negocios tienen variantes en WooCommerce,
  desde cuando llevan el stock desactualizado).
- Revisar si aplica el mismo patron de "struct propia que descarta campos nuevos
  silenciosamente" en otros consumers de colas del proyecto (Shopify, MercadoLibre),
  ya que es un riesgo estructural: cualquier campo nuevo agregado al mensaje del
  publisher requiere acordarse de replicarlo en CADA consumer que lo lea.
