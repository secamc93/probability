# WhatsApp por cliente: que falta y quien lo hace

Fecha: 2026-09-06
Rama: `claude/whatsapp-numero-por-cliente-s22bjf`
Contexto y decisiones: `.claude/project/whatsapp-numero-por-cliente.md`
Detalle tecnico: `back/central/services/integrations/messaging/whatsapp/NUMERO-PROPIO.md`
Pendientes criticos: `.claude/alerts/whatsapp-numero-propio-pendientes.md`

Este archivo es la lista de pasos, en orden, para no perder el hilo. Marcar con
`[x]` lo que se complete y anotar la fecha.

## El objetivo

Que cada negocio envie y reciba WhatsApp desde SU numero. Decidido el
2026-09-06: **por defecto el numero se agrega al WABA de Probability**, nosotros
pagamos las conversaciones a Meta y se las cobramos al negocio, y el numero
hereda las plantillas ya aprobadas. El WABA propio del cliente queda como opcion
para quien lo exija.

Hay dos formas de dar de alta el numero, y la segunda depende de Meta:

| | Asistente en nuestra pantalla | Registro insertado (Embedded Signup) |
|---|---|---|
| Que hace el cliente | digita su numero y el codigo que le llega | pulsa un boton e inicia sesion con Facebook |
| Codigo | listo | listo, apagado por flag |
| Depende de Meta | solo del cupo de numeros | App Review con Advanced Access |

## Bloque A - lo que hace el usuario en Meta

- [ ] **A1. Ampliar el cupo de numeros del negocio** en Meta Business Suite.
      Hoy la API responde `Phone Numbers Count Exceeded Limit Per Business`
      (subcodigo 2388386): estamos en el tope. **Sin esto no funciona ninguno de
      los dos caminos.**
- [x] **A2. Verificacion de acceso / proveedor de tecnologia.** Enviada el
      2026-09-06 22:32 por Sebastian Camacho. **Era la unica con fecha limite**
      (2026-11-05): sin ella Meta empieza a rechazar las llamadas de apps que
      usan otras empresas. Aplica al negocio, no a la app, y habilita
      `whatsapp_business_management` y `business_management` para clientes que
      no tengan rol en la app. Queda esperar la respuesta de Meta.
      Lo que opera hoy no corria peligro: el token de `cam-adm` esta emitido
      para la propia app y sobre el WABA del mismo negocio, y ese caso pasa el
      control. Vigilar la bandeja de alertas de la app y "Acciones requeridas".
- [ ] **A3. Subir el icono de la app** de 1024x1024 en Configuracion basica.
      Meta lo marca como faltante para poder enviar la solicitud.
- [ ] **A4. Elegir la categoria de la app** en Configuracion basica. Tambien
      esta marcada como faltante.
- [ ] **A4.1. Probar que `comercial@probabilityapp.com` recibe.** Es el correo
      que declaran la politica de privacidad, los terminos y la pagina de
      eliminacion de datos, y es donde escribe el revisor de Meta y donde llegan
      las solicitudes de datos de la Ley 1581. El dominio tiene MX
      (mx01/mx03.mi.com.co) y `gerencia@probabilityapp.com` esta en uso, pero de
      `comercial@` no hay rastro. Si rebota, hay que corregir los tres textos.
- [ ] **A5. Verificar el correo de contacto** (`probabilitysas@gmail.com`
      figura como no verificado).
- [ ] **A6. Crear la configuracion de registro insertado** en el panel
      (WhatsApp -> Registro insertado) y anotar el `config_id`.
- [ ] **A7. Grabar los screencast**, **uno por permiso** (son tres). Meta:
      "Any requested permission or feature missing a screen recording will not
      be approved", y cada uno con su descripcion propia, sin copiar y pegar.
      Cada video muestra: login desde sesion cerrada, el momento en que el
      cliente otorga el permiso (el Embedded Signup abriendose) y que hace la
      app con eso (mandar un mensaje real, gestionar plantillas). Forma: 1080p,
      pantalla de 1440 o menos, cursor grande, mouse en vez de teclado, **sin
      audio** (los revisores no lo escuchan) y **con subtitulos**, porque la
      interfaz esta en espanol. Sale solo cuando A1 este resuelto y haya un
      numero real conectado. Los guiones los puede redactar Claude.

- [ ] **A7.1. Hacer al menos una llamada real con cada permiso** antes de
      enviar, dentro de los 30 dias previos. Los dos de WhatsApp ya se usan a
      diario; **`business_management` no**, porque todavia no lo tenemos: hay
      que estrenarlo apenas la verificacion de proveedor lo habilite.

- [ ] **A7.2. Preparar el acceso del revisor**: un negocio demo con
      instrucciones paso a paso. Nunca la cuenta de un cliente real.
- [ ] **A8. Enviar la solicitud de App Review.** Ya esta armada y sin enviar,
      con los 4 permisos correctos: `whatsapp_business_messaging`,
      `whatsapp_business_management`, `business_management`, `public_profile`.
      **No tiene fecha limite.** Sin ella no se bloquea nada de lo que ya
      funciona; lo unico que no se puede es que un cliente sin rol en la app se
      conecte solo. Es un techo comercial, no una amenaza.

## Bloque B - codigo y despliegue

- [x] **B1. Paginas legales publicas** (2026-09-06). Tres paginas nuevas en el
      sitio Astro, que es el que sirve `/`, asi que no hubo que tocar nginx:
      `/politica-de-privacidad`, `/terminos-y-condiciones` y
      `/eliminacion-de-datos`. Las dos primeras reusan el HTML que ya existia en
      `legal_documents`; la tercera se escribio nueva. Enlazadas desde el footer
      y desde el formulario de registro, que apuntaban a `#`.
- [ ] **B2. Pegar las 3 URLs en Configuracion basica de la app.** Ojo: hoy
      Condiciones del servicio y Eliminacion de datos apuntan a
      `https://www.facebook.com/`, que es un placeholder y hace rechazar la
      revision. Requiere que B1 este desplegado.
- [ ] **B3. Merge y deploy de la rama.** 5 commits sin push.
- [ ] **B4. Correr las dos migraciones en produccion**:
      `migrateWhatsappInboundConversationType` (CHECK con `inbound`; sin ella
      toda conversacion entrante falla con 23514) y
      `migrateWhatsappPhoneNumberUnique` (indice unico del `phone_number_id`).
      Ya corrieron y se verificaron en local. Despues dejar `Migrate()` en cero.
- [ ] **B5. Cargar `waba_id` en las platform credentials de produccion**
      (`1302830408357767`). En la copia local no estaba; sin el, aprovisionar
      plantillas y resolver el WABA fallan.
- [ ] **B6. Dar el permiso `Integraciones-Mensajeria`** a los roles de negocio.
      Hoy el rol Administrador del negocio Demo no lo tiene, y por eso
      "Nueva Integracion" no ofrece la categoria Mensajeria. Confirmar primero
      como esta en produccion.
- [ ] **B7. Suscribir `phone_number_name_update`** en el webhook de la app. Hoy
      solo estan `messages` y `message_template_status_update`. Sin ese campo no
      nos enteramos de que Meta aprobo el nombre visible, y despues de esa
      aprobacion hay que volver a llamar `register`.
- [ ] **B8. Medidor de consumo y cobro.** Contar mensajes por `phone_number_id`
      y descontarlos de la billetera del negocio. **Sin esto, cada cliente que
      envie desde nuestro WABA es plata que Probability paga y no recupera.** El
      punto unico donde registrarlo es `sendTemplate`.
- [ ] **B9. Encender el registro insertado**: cargar en las platform credentials
      `embedded_signup_enabled`, `app_id`, `app_secret` (no esta en el repo) y
      el `embedded_signup_config_id` de A6.
- [x] **B10.1. Reorganizar el formulario** (2026-09-06). La pantalla sigue el
      diseno del formulario de Shopify (mismo ui-kit, hero + tarjetas, ancho
      angosto). Un switch decide todo: "Usar el numero predeterminado de
      Probability"; al apagarlo aparecen los dos caminos (agregar el numero a
      nuestra cuenta, o conectar la cuenta propia de Meta). Dentro del segundo,
      el registro insertado es la opcion principal y la conexion por
      credenciales queda plegada como avanzada, con sus requisitos escritos
      (admin de la WABA, permisos, y token de System User: uno de usuario
      caduca y corta los envios sin aviso).
      Se elimino la opcion que permitia reclamar un numero alojado escribiendo
      solo su `phone_number_id`, en la UI y en el backend.

- [ ] **B10. Exponer el `quality_rating` por numero en la UI.** Con N clientes
      son N calidades, y una sancion de politica de Meta es a nivel de WABA: si
      un cliente manda spam, caen todos los numeros alojados.

## Bloque C - pruebas que faltan

- [ ] **C1. Alta de un numero real** con el asistente, de punta a punta
      (agregar, codigo, verificar, activar). Depende de A1.
- [ ] **C2. Enviar y recibir** con ese numero: plantilla saliente y mensaje
      entrante enrutado al negocio dueno.
- [ ] **C3. Aprovisionar plantillas contra un WABA de cliente real.** La fase 3
      nunca corrio contra Meta. Solo aplica al camino de WABA propio.
- [ ] **C4. Registro insertado end to end**, con una cuenta con rol en la app.
      Depende de A6 y B9.
- [ ] **C5. Ticket del trabajo.** La regla `.claude/rules/tickets.md` lo pide y
      no se creo: el modulo exige JWT de super admin contra el backend.

## Que se comparte al alojar numeros en nuestro WABA

Decidido el 2026-09-06 y confirmado contra la API: `POST /{waba_id}/phone_numbers`
es un endpoint oficial y nuestro token ya tiene los permisos que pide (usuario o
System User con `whatsapp_business_management` + `whatsapp_business_messaging`).
Lo que hay que tener claro es que se comparte y que no:

| Que | Se comparte |
|---|---|
| Cupo de numeros por negocio | **Si**, y ya esta lleno (A1) |
| Factura de las conversaciones | **Si**: paga Probability, de ahi el medidor (B8) |
| Riesgo de sancion de politica (nivel WABA) | **Si**: un cliente que mande spam puede afectar a todos |
| Limite de mensajeria (1K/10K/100K por dia) | No, es por numero |
| Quality rating | No, es por numero |
| Plantillas | Si, y a favor: el numero nuevo hereda las 26 aprobadas |

El modelo que Meta prefiere a largo plazo es un WABA por cliente (creado "on
behalf of" via Embedded Signup), con la **linea de credito compartida** para que
la factura siga llegando a Probability. Eso exige ser Solution Partner con linea
otorgada por Meta, o sea volumen y acuerdo comercial: no es el primer paso.
Variante intermedia si un cliente es grande o riesgoso: crearle **su propio WABA
dentro de nuestro negocio**, con nuestro metodo de pago. Aisla la sancion y el
codigo de aprovisionamiento de plantillas ya sirve para eso.

## El orden que tiene sentido

1. **A1 (cupo)** desbloquea todo lo demas. Es lo primero.
2. **B3 + B4 + B5** para tener el codigo vivo en produccion.
3. **C1 y C2** con un numero real: ahi el producto ya se puede vender.
4. **B8 (medidor)** antes de conectar al segundo cliente, o la cuenta de Meta
   empieza a crecer sin contrapartida.
5. **A3 a A8 + B2** para App Review, en paralelo. El screencast de A7 sale de
   C1, asi que este bloque va despues de tener un cliente andando.
6. **B9 y C4** cuando Meta conceda Advanced Access.

## Lo que ya esta hecho (2026-09-05 y 06)

- Fases 1 a 5 del plan original: credenciales por negocio, ruteo del webhook por
  `phone_number_id`, plantillas por WABA, aislamiento de alertas y OTP, y UI.
- Correcciones de la revision del 2026-09-06: un negocio ya no puede declarar el
  numero de otro, el camino corto funciona con el token de Probability, y las
  plantillas de autenticacion se reconstruyen con la forma que Meta acepta.
- Alta del numero por API desde nuestra pantalla (4 pasos) y registro insertado
  completo detras de flag.
- Las dos migraciones, corridas y verificadas contra la copia local.
