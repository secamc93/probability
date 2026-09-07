# WhatsApp numero propio por cliente - pendientes

Fecha: 2026-09-05
Plan: `.claude/project/whatsapp-numero-por-cliente.md`
Detalle tecnico: `back/central/services/integrations/messaging/whatsapp/NUMERO-PROPIO.md`
Rama: `claude/whatsapp-numero-por-cliente-s22bjf`

El codigo de las fases 1 a 5 esta escrito y compila; `go build ./...` y
`go test ./...` de `back/central` pasan, y el front typechequea. Lo que sigue NO
se ejecuto: esta sesion no tiene base de datos, ni Redis, ni token de Meta.

## Revision del 2026-09-06

Se reviso la rama y se verifico contra Meta con el token de `cam-adm`. Lo que se
corrigio en esa pasada esta al final de este archivo.

## Urgente (antes de desplegar)

1. **Correr las dos migraciones en produccion.** `Migrate()` en
   `back/migration/internal/infra/repository/constructor.go` encadena
   `migrateWhatsappInboundConversationType` (CHECK de
   `whatsapp_conversations.conversation_type` con `inbound`; sin el, toda
   conversacion entrante falla con SQLSTATE 23514) y
   `migrateWhatsappPhoneNumberUnique` (indice unico parcial sobre
   `config->>'phone_number_id'`). **Ya corrieron y se verificaron contra la copia
   local** (2026-09-06); falta produccion. Al terminar, dejar `Migrate()` en cero
   y registrar la corrida en `back/migration/MIGRACIONES.md`.

2. **Probar la fase 1 contra la base local con el numero de test.** Un negocio
   sin `phone_number_id` propio debe seguir enviando por el numero de
   Probability (no romper a nadie), y uno con `phone_number_id` + `waba_id` +
   `access_token` debe enviar por el suyo. Es el caso que mas puede romper en
   produccion porque toca a TODOS los negocios de hoy.

3. **Poner `waba_id` en las credenciales de plataforma del integration_type 2.**
   Sin ese campo, `POST /whatsapp/templates/provision` no tiene de donde copiar
   las plantillas y devuelve error. Valor de produccion: `1302830408357767`.

## Importante

4. **Fase 3 sin verificar contra un WABA real.** El aprovisionamiento de
   plantillas y el consumo de `message_template_status_update` estan escritos
   pero nunca corrieron contra Meta. Necesitan un cliente conectado o un WABA de
   pruebas compartido con Probabilityapp. Lo que si esta verificado: el webhook
   de la app tiene suscrito `message_template_status_update` ademas de
   `messages`, y el WABA de produccion tiene 26 plantillas (no 13).

4.1 **Asignarle el WABA del cliente al system user `cam-adm`.** Su token tiene
   solo `whatsapp_business_management` y `whatsapp_business_messaging` (sin
   `business_management`), y `GET /me/assigned_whatsapp_business_accounts`
   devuelve vacio: los WABA compartidos no se pueden descubrir por API, hay que
   asignarlos a mano en Business Manager. Si falta ese paso, guardar la conexion
   falla con error de permisos de Meta, que es justo lo que se quiere.

5. **Quality rating por numero.** El riesgo que el propio plan senala sigue sin
   cubrir: hoy se vigila un `quality_rating` y con N clientes son N. La UI
   muestra el estado de plantillas, no la calidad del numero ni la ventana de
   24 h. Falta leer `GET /{phone_number_id}?fields=quality_rating` y exponerlo.

6. **Sin ticket.** La regla `.claude/rules/tickets.md` pide registrar el trabajo
   en un ticket y cerrarlo con el diagnostico. No se pudo: el modulo de tickets
   exige JWT de super admin contra un backend corriendo. Crear el ticket y pegar
   ahi el resumen de la rama.

## Deseable

7. **Embedded Signup.** El camino largo del plan sigue pendiente y depende de
   sacar la app de dev mode, publicar politica de privacidad / terminos /
   eliminacion de datos y pasar App Review con `business_management`. El codigo
   de las fases 1 a 5 sirve igual para ese camino; lo unico que cambia es la
   pantalla de conexion, que pasa a ser un boton.

## Cuando se cierra esta alerta

Cuando 1, 2 y 3 esten hechos y verificados, y 4 tenga al menos una corrida real
contra un WABA. Los puntos 5 a 7 se mueven a `ROADMAP.md` si siguen abiertos.

## Corregido el 2026-09-06

- **Un negocio podia quedarse con los mensajes de otro.** El formulario dejaba
  escribir cualquier `phone_number_id` y el ruteo del webhook le entregaba al
  declarante los entrantes de ese numero. Ahora `phone_number_id`, `waba_id` y
  `use_platform_token` son campos protegidos en el `PUT /integrations/:id`
  generico, y solo se escriben por `PUT /whatsapp/connection`, que rechaza el
  numero de la plataforma, rechaza uno ya tomado y verifica contra Meta que el
  numero pertenezca a ese WABA. Ademas hay indice unico en base.
- **El camino corto no estaba implementado.** Se exigia un `access_token` propio
  del cliente, que es justo el tramite con Meta que el plan queria evitar. Ahora,
  con `phone_number_id` + `waba_id` y sin credencial propia, se envia por el
  numero del negocio con el token de Probability. El token propio queda opcional.
- **`recuperacion_codigo` no se podia crear en el WABA del cliente.** Meta no
  acepta en el POST los `components` que devuelve el GET de una plantilla
  `AUTHENTICATION`. Se reconstruyen con la forma correcta, y ademas las
  plantillas que son de Probability hacia el negocio ya no se copian.
