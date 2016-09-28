// Copyright 2016 Stratumn SAS. All rights reserved.
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

package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/stratumn/go/dummystore"
	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/go/store/storehttp"
)

var (
	http     = flag.String("http", storehttp.DefaultAddress, "HTTP address")
	certFile = flag.String("tlscert", "", "TLS certificate file")
	keyFile  = flag.String("tlskey", "", "TLS private key file")
	version  = "0.1.0"
	commit   = "00000000000000000000000000000000"
)

func main() {
	flag.Parse()

	log.Infof("%s v%s@%s", dummystore.Description, version, commit[:7])
	log.Info("Copyright (c) 2016 Stratumn SAS")
	log.Info("Apache License 2.0")
	log.Infof("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	a := dummystore.New(&dummystore.Config{Version: version, Commit: commit})

	go func() {
		sigc := make(chan os.Signal)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigc
		log.WithField("signal", sig).Info("Got exit signal")
		log.Info("Stopped")
		os.Exit(0)
	}()

	c := &jsonhttp.Config{
		Address:  *http,
		CertFile: *certFile,
		KeyFile:  *keyFile,
	}
	h := storehttp.New(a, c)

	log.WithField("http", *http).Info("Listening")
	if err := h.ListenAndServe(); err != nil {
		log.WithField("error", err).Fatal("Server stopped")
	}
}
