package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"payment-service/internal/messaging"
	"payment-service/internal/repository"
	grpcTransport "payment-service/internal/transport/grpc"
	"payment-service/internal/usecase"

	pb "github.com/tolik-tachyon/proto-generated/paymentpb"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

func main() {
	db, err := sql.Open("postgres",
		"host=localhost port=5432 user=postgres password=Study.ollie dbname=payment_db sslmode=disable",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewPostgresPaymentRepository(db)

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	publisher, err := messaging.NewPublisher(nc)
	if err != nil {
		log.Fatal(err)
	}

	uc := usecase.NewPaymentUseCase(repo)
	uc.Publisher = publisher

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterPaymentServiceServer(
		grpcServer,
		grpcTransport.NewServer(uc),
	)

	log.Println("gRPC server running on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
