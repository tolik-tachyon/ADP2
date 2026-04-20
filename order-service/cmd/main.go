package main

import (
	"database/sql"
	"log"
	"os"

	"order-service/internal/repository"
	orderHTTP "order-service/internal/transport/http"
	"order-service/internal/usecase"

	pb "github.com/tolik-tachyon/proto-generated/paymentpb"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=Study.ollie dbname=order_db sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewPostgresOrderRepository(db)

	grpcAddr := os.Getenv("PAYMENT_GRPC_URL")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}

	conn, err := grpc.Dial(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	defer conn.Close()

	paymentClient := pb.NewPaymentServiceClient(conn)

	uc := usecase.NewOrderUseCase(repo, paymentClient)
	handler := orderHTTP.NewOrderHandler(uc)

	r := gin.Default()
	r.POST("/orders", handler.CreateOrder)
	r.GET("/orders/:id", handler.GetOrder)
	r.PATCH("/orders/:id/cancel", handler.CancelOrder)

	log.Println("Order service running on :8080")
	port := os.Getenv("ORDER_PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
