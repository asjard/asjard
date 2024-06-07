
# Generate *_rest.pb.go

```bash
go install github.com/asjard/protoc-gen-go-rest@latest
protoc --go-rest_out=${GOPATH}/src -I${GOPATH}/src -I. ./*.proto
```
