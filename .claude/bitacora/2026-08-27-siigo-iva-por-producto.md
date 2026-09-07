# Siigo: el IVA no se aplica automatico por producto, hay que declararlo por linea

## Resumen

Hoy la facturacion a Siigo usa un `tax_id` unico y fijo por negocio para
cualquier item con impuesto. Si un negocio (ej. Viga) tiene productos con
tarifas de IVA distintas (19%, 5%, exento), todos se facturan con la misma
tarifa. Se agrego (sin desplegar) la lectura del array `taxes[]` que Siigo
devuelve por producto, para poder validar si esto es un problema real y, si lo
es, aplicar el IVA correcto por SKU en vez de uno global.

## Contexto / pregunta que lo origino

El usuario pregunto si se puede consultar el IVA real configurado por producto
en Siigo para Viga, y si se puede usar eso en la facturacion.

## Como funciona hoy (verificado en codigo)

`back/central/services/integrations/invoicing/siigo/internal/infra/secondary/client/mappers/invoice.go`
(`BuildCreateInvoiceRequest`):

```go
if taxID > 0 && item.Tax > 0 {
    siigoItem.Taxes = []request.SiigoTax{{ID: taxID}}
}
```

`taxID` sale de `invoicing_configs.invoice_config.tax_id`, un solo valor por
negocio. `item.Tax` solo se usa como bandera (>0), no aporta la tarifa.

**Siigo NO aplica automaticamente el impuesto que el producto tiene configurado
en su ficha maestra al crear una factura.** El payload de creacion solo manda
`items[].code` para identificar el producto; si esa linea no trae
`items[].taxes[]` explicito, Siigo la factura sin impuesto, sin importar la
configuracion del producto en su catalogo. Es la razon por la que existe el
`tax_id` fijo: sin el, ninguna factura llevaria IVA.

## Que se agrego (solo lectura, sin desplegar)

Antes, el listado/busqueda de productos de Siigo (`/v1/products`) descartaba el
campo `taxes` de la respuesta. Se agrego su captura:

- `domain/dtos/operation_types.go`: `ProductItem.Taxes []ProductTax`
  (`ID, Name, Type, Percentage`).
- `infra/secondary/client/list_products.go` y `search_products.go`: parsean
  `results[].taxes[]` de la respuesta de Siigo y lo mapean.
- `domain/dtos/catalog_types.go`: `CatalogItem.Taxes` para exponerlo en el
  endpoint `GET /siigo/products/search`.

Con esto se puede consultar, por SKU, que `tax_id`/porcentaje tiene cada
producto de Viga en Siigo. **Todavia no se corrio la consulta contra los datos
reales de Viga** (pendiente el deploy).

## Para aplicar IVA por producto en la facturacion (si los datos lo justifican)

No implementado aun. Haria falta:

1. Consultar el catalogo real de Viga y confirmar si de verdad hay tarifas
   mixtas (si todo esta al mismo IVA, no vale la pena la complejidad).
2. Al armar cada factura, resolver el `tax_id` por `code`/SKU contra el
   catalogo de Siigo, con cache (no se puede golpear la API de Siigo en cada
   factura).
3. Reemplazar el `taxID` fijo de `BuildCreateInvoiceRequest` por el que
   corresponda a cada item.

## Pendientes

- Desplegar la lectura de `taxes[]` y consultar el catalogo real de Viga.
- Decidir con el usuario si vale la pena implementar el IVA por producto o si
  el `tax_id` unico es suficiente en la practica.
- No confundir con el bug ya cerrado de `invalid_total_payments`
  (`2026-08-27-siigo-documento-inactivo-y-errores-crudos.md`): ese era sobre
  el monto del pago, no sobre el IVA.
