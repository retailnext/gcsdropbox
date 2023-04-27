// Copyright 2020 RetailNext, Inc.
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
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"golang.org/x/term"
	"cloud.google.com/go/iam/credentials/apiv1/credentialspb"
)

func setupLogger() func() {
	var logger *zap.Logger
	var err error
	if term.IsTerminal(int(os.Stdin.Fd())) {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)

	return func() {
		_ = logger.Sync()
	}
}

func setupInterruptContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case sig := <-c:
			zap.S().Infow("shutting_down", "signal", sig)
			cancel()
		case <-ctx.Done():
		}
	}()
	onExit := func() {
		signal.Stop(c)
		cancel()
	}
	return ctx, onExit
}

var (
	bucketFlag, nameFlag, methodFlag, serviceAccountFlag string
	credsClient                                          *credentials.IamCredentialsClient
	expirationFlag                                       time.Duration
)

func init() {
	flag.DurationVar(&expirationFlag, "expiration", 1*time.Hour, "Link expiration (max 12 hours)")
	flag.StringVar(&bucketFlag, "bucket", "", "GCS bucket name")
	flag.StringVar(&methodFlag, "method", "PUT", "HTTP method (PUT or GET)")
	flag.StringVar(&nameFlag, "name", "", "Name of the object (inside bucket)")
	flag.StringVar(&serviceAccountFlag, "serviceAccount", "", "Service account email address")
}

func main() {
	flag.Parse()
	if bucketFlag == "" || nameFlag == "" || serviceAccountFlag == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	sync := setupLogger()
	defer sync()
	lgr := zap.S()

	ctx, onExit := setupInterruptContext()
	defer onExit()

	var err error
	credsClient, err = credentials.NewIamCredentialsClient(ctx)
	if err != nil {
		lgr.Fatalw("credentials_client_error", "err", err)
	}

	var url string
	url, err = storage.SignedURL(bucketFlag, nameFlag, &storage.SignedURLOptions{
		Method:         methodFlag,
		GoogleAccessID: serviceAccountFlag,
		Expires:        time.Now().Add(expirationFlag),
		SignBytes:      signBytes(ctx, serviceAccountFlag),
	})
	if err != nil {
		lgr.Fatalw("sign_url_error", "err", err)
	}

	fmt.Println(url)
}

func signBytes(ctx context.Context, serviceAccountID string) func([]byte) ([]byte, error) {
	return func(b []byte) ([]byte, error) {
		resp, err := credsClient.SignBlob(ctx, &credentialspb.SignBlobRequest{
			Payload: b,
			Name:    serviceAccountID,
		})
		if err != nil {
			return nil, err
		}
		return resp.SignedBlob, nil
	}
}
