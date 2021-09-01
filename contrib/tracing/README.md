# Tracing
With [Open tracing](https://opentracing.io/) and [Jaeger](https://www.jaegertracing.io/)

* Start Jaeger
```shell
docker run --name=jaeger -d -p 6831:6831/udp -p 16686:16686 jaegertracing/all-in-one
```
* Open UI

[http://localhost:16686](http://localhost:16686)

## Tgrade
Run with `--twasm.open-tracing` flag
```shell
tgrade start --twasm.open-tracing
```