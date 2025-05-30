//go:build linux && !386

// Copyright 2013-2015 go-diameter authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package diam_test

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/rakeshgmtke/go-diameter/diam"
	"github.com/rakeshgmtke/go-diameter/diam/diamtest"
)

func TestCapabilitiesExchangeSCTP(t *testing.T) {
	errc := make(chan error, 1)

	smux := diam.NewServeMux()
	smux.Handle("CER", handleCER(errc, false))

	srv := diamtest.NewServerNetwork("sctp", smux, nil)
	defer srv.Close()

	wait := make(chan struct{})
	cmux := diam.NewServeMux()
	cmux.Handle("CEA", handleCEA(errc, wait))

	cli, err := diam.DialNetwork("sctp", srv.Addr, cmux, nil)
	if err != nil {
		t.Fatal(err)
	}

	sendCER(cli)

	select {
	case <-wait:
	case err := <-errc:
		t.Fatal(err)
	case err := <-smux.ErrorReports():
		t.Fatal(err)
	case <-time.After(time.Second):
		t.Fatal("Timed out: no CER or CEA received")
	}
}

func TestCapabilitiesExchangeSCTP_TLS(t *testing.T) {
	errc := make(chan error, 1)

	cert, err := tls.LoadX509KeyPair("testdata/example-cert.pem", "testdata/example-key.pem")
	if err != nil {
		t.Fatal(err)
	}

	smux := diam.NewServeMux()
	smux.Handle("CER", handleCER(errc, true))

	srv := diamtest.NewUnstartedServerNetwork("sctp", smux, nil)
	tm := 100 * time.Millisecond
	srv.Config.ReadTimeout = tm
	srv.Config.WriteTimeout = tm
	srv.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}
	srv.StartTLS()
	time.Sleep(time.Millisecond * 10) // let srv start
	defer srv.Close()

	wait := make(chan struct{})
	cmux := diam.NewServeMux()
	cmux.Handle("CEA", handleCEA(errc, wait))

	cli, err := diam.DialNetworkTLS("sctp", srv.Addr, "testdata/example-cert.pem", "testdata/example-key.pem", cmux, nil)
	if err != nil {
		t.Fatal(err)
	}

	sendCER(cli)

	select {
	case <-wait:
	case err := <-errc:
		t.Fatal(err)
	case <-time.After(time.Second):
		t.Fatal("Timed out: no CER or CEA received")
	}
}
