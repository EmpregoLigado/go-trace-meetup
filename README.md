# Distributed Tracing

Repositório com exemplo de tracing em aplicações Go usando Jaeger e opencensus.

Apresentação em:

[https://levee.ml/lvX](https://levee.ml/lvX)

## Executando o Jaeger via Docker.
```sh
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.6
```

Acesso à interface web em [http://localhost:16686](http://localhost:16686)

## Executando exemplo

```
git clone git@github.com:EmpregoLigado/go-trace-meetup.git
cd go-trace-meetup
go mod tidy
go run main.go
```

Em um outro terminal ou em seu browser acesse:

```
curl http://0.0.0.0:8080/fib?rank=3
```

Visualize os traces em [http://localhost:16686](http://localhost:16686)
