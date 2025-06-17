package main

import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    "log"
)

func loadTLS(certFile, keyFile, caFile string) *tls.Config {
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        log.Fatalf("Error loading cert/key: %v", err)
    }

    caCert, err := ioutil.ReadFile(caFile)
    if err != nil {
        log.Fatalf("Error reading CA file: %v", err)
    }

    caPool := x509.NewCertPool()
    caPool.AppendCertsFromPEM(caCert)

    return &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caPool,
        // ClientAuth:   tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: true,
    }
}
