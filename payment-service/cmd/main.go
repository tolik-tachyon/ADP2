package main

import (
	"database/sql"
	"log"
	"net"

	"payment-service/internal/repository"
	grpcTransport "payment-service/internal/transport/grpc"
	"payment-service/internal/usecase"

	pb "github.com/tolik-tachyon/proto-generated/paymentpb"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	db, err := sql.Open("postgres",
		"host=localhost port=5432 user=postgres password=Study.ollie dbname=payments_db sslmode=disable",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewPostgresPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)

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
