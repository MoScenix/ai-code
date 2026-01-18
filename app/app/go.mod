module github.com/MoScenix/ai-code/app/app

go 1.25.5

replace github.com/apache/thrift => github.com/apache/thrift v0.13.0

require (
github.com/golang/protobuf v1.5.4 // indirect
github.com/MoScenix/ai-code/common => ../../common
	github.com/MoScenix/ai-code/rpc_gen => ../../rpc_gen
    )
