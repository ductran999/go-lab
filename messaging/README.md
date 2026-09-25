# Messaging — async compared side by side

One CLI benchmarking Kafka client libraries (plus RabbitMQ):
same produce workload, four implementations. Brought in from
`xplr-distributed-mq`; imports rewritten, binaries to `./bin`.

## Map

| Dir | What |
|-----|------|
| `evtstream/kafka/{confluent,franzgo,kafkago,sarama}/` | Producers per client lib |
| `examples/producer/{kafka,rabbitmq}/` | Runnable produce examples |
| `mq/rabbitmq/` | AMQP producer |
| `cmd/` | Cobra CLI (`kafka`, version) |

## Run

```bash
make -C messaging kafka  # broker + kafka-ui (from repo root: docker compose)
go run ./messaging --help
```

Needs Kafka/RabbitMQ up (`docker-compose.yaml`). Legacy demos:
runnable (build/vet/test) but exempt from strict style.
