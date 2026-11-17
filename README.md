# go-counterpass

## Dependencias y compilación

Requiere Go 1.25 o posterior. Las dependencias se fijan en `go.mod` y
`go.sum`; NATS y Proto.Actor usan los forks publicados en
`github.com/dumacp/nats.go` y `github.com/dumacp/protoactor-go`, con los
cambios propios del proyecto. Los `replace` de estos forks deben mantenerse
en el módulo principal: Go no hereda los `replace` de sus dependencias.

`go-params` usa la versión `v1.0.4`; `go-optocontrol` y `sonar` usan `v0.1.0`,
que incluyen las interfaces de los contadores utilizadas por este proyecto.

Al revisar el acceso el 17 de noviembre de 2026, `dumacp/go-doors` y
`dumacp/go-params` requerían autenticación en GitHub. Para descargarlos se
necesita una cuenta autorizada y configurar `GOPRIVATE` para esos módulos,
conservando cualquier configuración privada adicional del entorno.

`go-logirastreo` usa la versión publicada `v0.1.0`. No se necesitan
directorios hermanos ni reemplazos a rutas locales para compilar.

Comprobación de compilación, sin ejecutar las pruebas de puerto serie:

```sh
GOWORK=off go test ./... -run '^$'
GOWORK=off go build ./main
```

Compilación para Linux ARMv7:

```sh
GOWORK=off CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
  go build -trimpath -ldflags='-s -w' -o /tmp/go-counterpass-armv7 ./main
```

Las pruebas existentes `TestNewCountingActor` y `Test_tamperingFlipFlop`
pasan en ejecución normal. Con `-race`, `TestNewCountingActor` detecta
accesos concurrentes en `internal/pubsub/pubsub.go` (líneas 89 y 133).
El mismo fallo se reproduce usando las dependencias locales anteriores a
esta actualización; queda pendiente corregirlo por separado.
