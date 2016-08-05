// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

// dummystore starts HTTP server with a dummystore.
//
// Usage
//
//	-port string
//		server port (default ":5000")
//	-tlscert string
//		TLS certificate file
//	-tlskey string
//		TLS private key file
//	-verbose
//	  	verbose output
//
// Docker
//
//	docker run -p 5000:5000 stratumn/dummystore
package main