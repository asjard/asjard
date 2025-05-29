/*
 *
 * Copyright 2024 ASJARD authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// protoc-gen-go-rabbitmq is a plugin for the Google protocol buffer compiler to
// generate Go code. Install it by building this program and making it
// accessible within your PATH with the name:
//
//	protoc-gen-go-rabbitmq
//
// The 'go-rabbitmq' suffix becomes part of the argument for the protocol compiler,
// such that it can be invoked as:
//
//	protoc --go-rabbitmq_out=. path/to/file.proto
//
// This generates Go service definitions for the protocol buffer defined by
// file.proto.  With that input, the output will be written to:
//
//	path/to/file_rabbitmq.pb.go
package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	version = "1.0.0"
	name    = "protoc-gen-go-rabbitmq"
)

var requireUnimplemented *bool
var useGenericStreams *bool
var flags flag.FlagSet

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("%s %v\n", name, version)
		return
	}

	var flags flag.FlagSet
	requireUnimplemented = flags.Bool("require_unimplemented_servers", true, "set to false to match legacy behavior")
	useGenericStreams = flags.Bool("use_generic_streams_experimental", false, "set to true to use generic types for streaming client and server objects; this flag is EXPERIMENTAL and may be changed or removed in a future release")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			NewRabbitmqGenerator(gen, f).Run()
		}
		return nil
	})
}
