package helper

import (
	"context"
	"errors"
	"flag"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

var (
	tls    = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile = flag.String("ca_file", "/key/ca.pem", "")
)

func GRPCConnection(serverUrl string) (*grpc.ClientConn, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer ctxCancel()
	if serverUrl == "" {
		return nil, errors.New("server Url not passed")
	}
	var opts []grpc.DialOption
	if *tls {
		if *caFile == "" {
			*caFile = testdata.Path("ca.pem")
		}
		splitVal := strings.Split(serverUrl, ":")
		cred, err := credentials.NewClientTLSFromFile(*caFile, splitVal[0])
		if err != nil {
			log.Errorf("Failed to create TLS credentials %v", err)
			return nil, err

		}
		log.Info("Successfully created TLS credentials")
		opts = append(opts, grpc.WithTransportCredentials(cred))
		//opts = append(opts, grpc.WithKeepaliveParams(kacp))
	} else {
		log.Info("creates with insecure")
		opts = append(opts, grpc.WithInsecure())
	}
	log.Info("Dialing ...")
	conn, err := grpc.DialContext(ctx, serverUrl, opts...)
	if err != nil {
		log.Errorf("fail to dial: %v", err)
		return nil, err
	}
	log.Info("GRPC Connected")
	return conn, nil
}
func CloseGRPCConnection(conn grpc.ClientConnInterface) {
	err := conn.(*grpc.ClientConn).Close()
	if err != nil {
		log.Error("gRpc Connection is not closed")
	}
	log.Info("gRpc Connection is closed")
}
