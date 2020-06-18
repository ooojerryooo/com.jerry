package main

import (
	pb "com.jerry/grpc-test/server/proto/helloworld"
	"context"
	"crypto/tls"
	"crypto/x509"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"os"
)

const (
	address = "localhost:50002"
	//address = "172.20.48.107:7051"
)

func main() {
	//creds, err := credentials.NewClientTLSFromFile("grpc-test/conf/server.crt", "ewklvehr-out")
	//if err != nil {
	//	panic(fmt.Errorf("could not load tls cert: %s", err))
	//}

	//creds := credentials.NewTLS(&tls.Config{
	//	InsecureSkipVerify: true,
	//})

	//cert, err := tls.LoadX509KeyPair("grpc-test/conf/client.crt", "grpc-test/conf/client.key")
	//if err != nil {
	//	log.Fatalf("tls.LoadX509KeyPair err: %v", err)
	//}


	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("grpc-test/conf/ca.crt")
	if err != nil {
		log.Fatalf("ioutil.ReadFile err: %v", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("certPool.AppendCertsFromPEM err")
	}

	creds := credentials.NewTLS(&tls.Config{
		//Certificates: []tls.Certificate{cert},
		ServerName:   "ewklvehr-out",
		RootCAs:      certPool,
	})



	//conn, err := grpc.Dial(address, grpc.WithInsecure())
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	name := "lin"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	r, err := c.SayHello(context.Background(), &pb.HelloRequest{Name: name})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Println(r.Message)
}
