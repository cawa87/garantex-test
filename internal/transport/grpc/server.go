package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/cawa87/garantex-test/internal/lib/logger/sl"
	"github.com/cawa87/garantex-test/internal/repository/postgres"
	"github.com/cawa87/garantex-test/internal/service/exchange"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/cawa87/garantex-test/gen/go/rate_service.v1"
)

type Server struct {
	pb.UnimplementedRateServiceServer
	repo     *postgres.Repository
	exchange *exchange.Client
	logger   *sl.Logger
}

func NewServer(repo *postgres.Repository, exchange *exchange.Client, logger *sl.Logger) *Server {
	return &Server{
		repo:     repo,
		exchange: exchange,
		logger:   logger,
	}
}

func (s *Server) GetRates(ctx context.Context, req *pb.GetRatesRequest) (*pb.GetRatesResponse, error) {
	s.logger.Info("GetRates called")

	rate, err := s.exchange.GetRates(ctx)
	if err != nil {
		s.logger.Error("Failed to get rates from exchange", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get rates from exchange: %v", err)
	}

	if err := s.repo.SaveRate(ctx, rate); err != nil {
		s.logger.Error("Failed to save rate to database", "error", err)
	}

	response := &pb.GetRatesResponse{
		Ask:       rate.Ask,
		Bid:       rate.Bid,
		Timestamp: timestamppb.New(rate.Timestamp),
	}

	s.logger.Info("GetRates completed successfully",
		"ask", response.Ask,
		"bid", response.Bid)

	return response, nil
}

func (s *Server) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	s.logger.Debug("HealthCheck called")

	dbCount, err := s.repo.GetRatesCount(ctx)
	if err != nil {
		s.logger.Error("Database health check failed", "error", err)
		return &pb.HealthCheckResponse{
			Status: "unhealthy",
			Details: map[string]string{
				"database": "connection failed",
				"error":    err.Error(),
			},
		}, nil
	}

	exchangeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err = s.exchange.GetRates(exchangeCtx)
	exchangeStatus := "healthy"
	if err != nil {
		s.logger.Warn("Exchange health check failed", "error", err)
		exchangeStatus = "unhealthy"
	}

	overallStatus := "healthy"
	if exchangeStatus == "unhealthy" {
		overallStatus = "degraded"
	}

	response := &pb.HealthCheckResponse{
		Status: overallStatus,
		Details: map[string]string{
			"database":         "healthy",
			"database_records": fmt.Sprintf("%d", dbCount),
			"exchange":         exchangeStatus,
			"timestamp":        time.Now().Format(time.RFC3339),
		},
	}

	s.logger.Debug("HealthCheck completed", "status", overallStatus)
	return response, nil
}

func (s *Server) Run(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(s.loggingInterceptor()),
	)
	pb.RegisterRateServiceServer(grpcServer, s)

	s.logger.Info("Starting gRPC server", "port", port)

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) loggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		s.logger.Info("gRPC request started",
			"method", info.FullMethod,
			"request", fmt.Sprintf("%T", req))

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			s.logger.Error("gRPC request failed",
				"method", info.FullMethod,
				"duration", duration,
				"error", err)
		} else {
			s.logger.Info("gRPC request completed",
				"method", info.FullMethod,
				"duration", duration)
		}

		return resp, err
	}
}
