inc1 := ^TestIteration1$
inc2 := ^TestIteration2$
inc3 := ^TestIteration3$
inc4 := ^TestIteration4$

build: build-agent build-server

build-agent:
	go build  -o ./cmd/agent/agent ./cmd/agent && chmod +x ./cmd/agent/agent
build-server:
	go build  -o ./cmd/server/server ./cmd/server && chmod +x ./cmd/server/server

tests: build tests-inc-1 tests-inc-2 tests-inc-3 tests-inc-4

tests-inc-1:
	./devopstest -test.v -test.run=$(inc1) -source-path=. -binary-path=./cmd/server/server -agent-binary-path=./cmd/agent/agent
tests-inc-2:
	./devopstest -test.v -test.run=$(inc2) -source-path=. -binary-path=./cmd/server/server -agent-binary-path=./cmd/agent/agent
tests-inc-3:
	./devopstest -test.v -test.run=$(inc3) -source-path=. -binary-path=./cmd/server/server -agent-binary-path=./cmd/agent/agent
tests-inc-4:
	./devopstest -test.v -test.run=$(inc4) -source-path=. -binary-path=./cmd/server/server -agent-binary-path=./cmd/agent/agent