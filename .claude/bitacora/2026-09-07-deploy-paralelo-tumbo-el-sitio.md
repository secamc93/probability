# 502 en produccion: dos deploys pisandose el archivo de upstreams

Fecha: 2026-09-07
Ticket: pendiente de crear
Reglas relacionadas: `.claude/rules/deploy.md`

## Que paso

El merge del PR #93 (WhatsApp: numero propio por cliente) toco backend y
frontend, asi que se dispararon los dos workflows a la vez. Los cuatro deploys
terminaron en verde, pero el sitio quedo caido: todo `/api/` devolvia 502 y el
login mostraba en pantalla `Unexpected token '<', "<html> <h"... is not valid
JSON`, porque el front recibia la pagina de error de nginx donde esperaba JSON.

## Causa raiz

`active.conf` guarda los dos upstreams en un solo archivo. Cada script de deploy
leia el color del OTRO servicio al empezar y reescribia el archivo ENTERO al
final. Entre esas dos cosas pasan minutos.

Los dos deploys escribieron con 0,2 s de diferencia:

| Hora | Deploy | Escribio |
|---|---|---|
| 05:01:52.73 | Frontend | `backend=blue frontend=blue` |
| 05:01:52.94 | Backend | `backend=green frontend=green` |

Gano el ultimo en escribir segun el orden real de los `mv`, y quedo
`backend=blue` cuando blue ya estaba apagado. nginx enrutando a un contenedor
inexistente.

Estado encontrado en el servidor:

```
central_reserve_prod_green   Up 20 minutes (healthy)
frontend_prod_blue           Up 20 minutes (healthy)
active.conf -> back-central-blue   <-- no existe
```

## Como se corrigio

1. Mitigacion inmediata: `sed` sobre `active.conf` cambiando el upstream del
   backend a green, `nginx -t` y `nginx -s reload`. El sitio volvio de una.
2. Correccion de fondo: `bg_switch <servicio> <color>` en `bluegreen-lib.sh`,
   con candado por `mkdir`, relectura del color del otro servicio dentro de la
   seccion critica, escritura atomica con `mv` y restauracion si `nginx -t`
   falla. Los dos deploys llaman a esa funcion en vez de escribir el archivo
   entero.

## Verificacion

Se reprodujo el incidente con `dash` y stubs de docker/nginx: con la version
vieja el segundo deploy pisaba el color del primero; con `bg_switch` el frontend
conserva `backend=green`. Tambien se lanzaron dos switches simultaneos: uno
espera al otro ("Otro deploy esta cambiando upstreams"), el resultado final es
correcto y no queda candado huerfano.

## Hipotesis descartadas

- **No fue el codigo de WhatsApp.** El backend green estaba arriba y healthy;
  respondia `/health` por dentro. El problema era a quien apuntaba nginx.
- **No fue falta de las migraciones.** Ninguna integracion tiene numero propio
  todavia, asi que el CHECK de `inbound` no se toco.
- **No fue iptables.** El sitio respondia; era nginx devolviendo 502 de un
  upstream inexistente, no un problema de red.

## Lo que ya estaba escrito y no alcanzo

`.claude/rules/deploy.md` decia "no desplegar backend y frontend a la vez", pero
como recomendacion para humanos. Un solo push que toca los dos lo dispara solo,
que es el caso normal de un PR de producto.
