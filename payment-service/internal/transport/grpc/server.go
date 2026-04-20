package grpc

import (
	"context"
	"time"

	"payment-service/internal/domain"
	"payment-service/internal/usecase"

	pb "github.com/tolik-tachyon/proto-generated/paymentpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedPaymentServiceServer
	UseCase *usecase.PaymentUseCase
}

func NewServer(uc *usecase.PaymentUseCase) *Server {
	return &Server{UseCase: uc}
}

func (s *Server) ProcessPayment(
	ctx context.Context,
	req *pb.PaymentRequest,
) (*pb.PaymentResponse, error) {

	payment := &domain.Payment{
		OrderID: req.OrderId,
		Amount:  req.Amount,
	}

	err := s.UseCase.AuthorizePayment(payment)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PaymentResponse{
		Status:        payment.Status,
		TransactionId: payment.TransactionID,
	}, nil
}

func (s *Server) SubscribePaymentStatus(
	req *pb.PaymentRequest,
	stream pb.PaymentService_SubscribePaymentStatusServer,
) error {

	_ = stream.Send(&pb.PaymentResponse{
		Status: "Pending",
	})

	time.Sleep(1 * time.Second)

	payment := &domain.Payment{
		OrderID: req.OrderId,
		Amount:  req.Amount,
	}

	_ = s.UseCase.AuthorizePayment(payment)

	_ = stream.Send(&pb.PaymentResponse{
		Status:        payment.Status,
		TransactionId: payment.TransactionID,
	})

	return nil
}

func (s *Server) ListPayment(
	ctx context.Context,
	req *pb.ListPaymentRequest,
) (*pb.ListPaymentResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
