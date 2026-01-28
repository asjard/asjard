## build protobuf

```bash
make gen_proto
```

## Run service

```bash
# run svc-gw
services=svc-gw/apis/api make run_dev
# run svc-example
services=svc-example/apis/api make run_dev
```
