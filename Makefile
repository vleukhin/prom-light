inc1 := ^TestIteration1$
inc2 := ^TestIteration2$
inc3 := ^TestIteration3$
inc4 := ^TestIteration4$
inc5 := ^TestIteration5$
inc6 := ^TestIteration6$
inc7 := ^TestIteration7$
inc8 := ^TestIteration8$
inc9 := ^TestIteration9$
inc10 := ^TestIteration10$
inc11 := ^TestIteration11$
inc12 := ^TestIteration12$
inc13 := ^TestIteration13$
inc14 := ^TestIteration14$
curr_dir := $(PWD)

build: build-agent build-server build-staticlint

build-agent:
	go build  -o ./cmd/agent/agent ./cmd/agent && chmod +x ./cmd/agent/agent
build-server:
	go build  -o ./cmd/server/server ./cmd/server && chmod +x ./cmd/server/server
build-staticlint:
	go build  -o ./cmd/staticlint/staticlint ./cmd/staticlint && chmod +x ./cmd/staticlint/staticlint
lint:
	docker run --rm -v $(curr_dir):/app -w /app golangci/golangci-lint:v1.45.2 golangci-lint run -v

proto-gen:
	protoc --go_out=. --go_opt=paths=source_relative \
      --go-grpc_out=. --go-grpc_opt=paths=source_relative \
      ./internal/proto/metrics.proto

tests: build tests-inc-1 tests-inc-2 tests-inc-3 tests-inc-4 tests-inc-5 tests-inc-6 tests-inc-7 tests-inc-8

tests-inc-1:
	./devopstest -test.v -test.run=$(inc1) -source-path=. -binary-path=./cmd/server/server -agent-binary-path=./cmd/agent/agent
tests-inc-2:
	./devopstest -test.v -test.run=$(inc2) -source-path=. -binary-path=./cmd/server/server -agent-binary-path=./cmd/agent/agent
tests-inc-3:
	./devopstest -test.v -test.run=$(inc3) -source-path=. -binary-path=./cmd/server/server -agent-binary-path=./cmd/agent/agent
tests-inc-4:
	./devopstest -test.v -test.run=$(inc4) -source-path=. -binary-path=./cmd/server/server -agent-binary-path=./cmd/agent/agent
tests-inc-5:
	./devopstest -test.v -test.run=$(inc5) -source-path=. -binary-path=./cmd/server/server -agent-binary-path=./cmd/agent/agent -server-port=4588
tests-inc-6:
	./devopstest -test.v -test.run=$(inc6) \
	-source-path=. \
 	-binary-path=./cmd/server/server \
 	-agent-binary-path=./cmd/agent/agent \
 	-server-port=4588 \
 	-database-dsn='postgres://postgres:postgres@localhost:5454/praktikum?sslmode=disable' \
 	-file-storage-path=/tmp/devops-metrics-db-test.json
tests-inc-7:
	./devopstest -test.v -test.run=$(inc7) \
	-source-path=. \
 	-binary-path=./cmd/server/server \
 	-agent-binary-path=./cmd/agent/agent \
 	-server-port=4588 \
 	-database-dsn='postgres://postgres:postgres@localhost:5454/praktikum?sslmode=disable' \
 	-file-storage-path=/tmp/devops-metrics-db-test.json
tests-inc-8:
	./devopstest -test.v -test.run=$(inc8) \
	-source-path=. \
 	-binary-path=./cmd/server/server \
 	-agent-binary-path=./cmd/agent/agent \
 	-server-port=4588 \
 	-database-dsn='postgres://postgres:postgres@localhost:5454/praktikum?sslmode=disable' \
 	-file-storage-path=/tmp/devops-metrics-db-test.json
tests-inc-9:
	./devopstest -test.v -test.run=$(inc9) \
	-source-path=. \
 	-binary-path=./cmd/server/server \
 	-agent-binary-path=./cmd/agent/agent \
 	-server-port=4588 \
 	-database-dsn='postgres://postgres:postgres@localhost:5454/praktikum?sslmode=disable' \
 	-file-storage-path=/tmp/devops-metrics-db-test.json \
 	-key="super-secret-key"
tests-inc-10:
	./devopstest -test.v -test.run=$(inc10) \
	-source-path=. \
 	-binary-path=./cmd/server/server \
 	-agent-binary-path=./cmd/agent/agent \
 	-server-port=4588 \
 	-database-dsn='postgres://postgres:postgres@localhost:5454/praktikum?sslmode=disable' \
 	-file-storage-path=/tmp/devops-metrics-db-test.json \
 	-key="super-secret-key"
tests-inc-11:
	./devopstest -test.v -test.run=$(inc11) \
	-source-path=. \
 	-binary-path=./cmd/server/server \
 	-agent-binary-path=./cmd/agent/agent \
 	-server-port=4588 \
 	-database-dsn='postgres://postgres:postgres@localhost:5454/praktikum?sslmode=disable' \
 	-file-storage-path=/tmp/devops-metrics-db-test.json \
 	-key="super-secret-key"
tests-inc-12:
	./devopstest -test.v -test.run=$(inc12) \
	-source-path=. \
 	-binary-path=./cmd/server/server \
 	-agent-binary-path=./cmd/agent/agent \
 	-server-port=4588 \
 	-database-dsn='postgres://postgres:postgres@localhost:5454/praktikum?sslmode=disable' \
 	-file-storage-path=/tmp/devops-metrics-db-test.json \
 	-key="super-secret-key"
tests-inc-13:
	./devopstest -test.v -test.run=$(inc13) \
	-source-path=. \
 	-binary-path=./cmd/server/server \
 	-agent-binary-path=./cmd/agent/agent \
 	-server-port=4588 \
 	-database-dsn='postgres://postgres:postgres@localhost:5454/praktikum?sslmode=disable' \
 	-file-storage-path=/tmp/devops-metrics-db-test.json \
 	-key="super-secret-key"
tests-inc-14:
	./devopstest -test.v -test.run=$(inc14) \
	-source-path=. \
 	-binary-path=./cmd/server/server \
 	-agent-binary-path=./cmd/agent/agent \
 	-server-port=4588 \
 	-database-dsn='postgres://postgres:postgres@localhost:5454/praktikum?sslmode=disable' \
 	-file-storage-path=/tmp/devops-metrics-db-test.json \
 	-key="super-secret-key"