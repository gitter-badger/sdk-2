// Copyright 2017 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The command postgresstore starts an HTTP server with a postgresstore.
package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/stratumn/sdk/postgresstore"
	"github.com/stratumn/sdk/store/storehttp"
)

var (
	version = "x.x.x"
	commit  = "00000000000000000000000000000000"
)

func init() {
	storehttp.RegisterFlags()
	postgresstore.RegisterFlags()
}

func main() {
	flag.Parse()

	log.Infof("%s v%s@%s", postgresstore.Description, version, commit[:7])

	a := postgresstore.InitializeWithFlags(version, commit)
	storehttp.RunWithFlags(a)
}
