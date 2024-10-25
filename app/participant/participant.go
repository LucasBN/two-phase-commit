package participant

import (
	"context"
	"log"
	"net"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "lucasbn/two-phase-commit/app/participant/proto"
)

// Server is used to implement the Greeter service
type Server struct {
	pb.UnimplementedTransactionServiceServer
	pool *pgxpool.Pool
	txs  map[string]pgx.Tx
}

func (s *Server) Begin(ctx context.Context, _ *pb.BeginRequest) (*pb.BeginReply, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// Generate a unique transaction ULID
	txID := ulid.Make().String()

	// Store the transaction in the txs map
	s.txs[txID] = tx

	return &pb.BeginReply{TxId: txID}, nil
}

// Query
func (s *Server) Query(ctx context.Context, req *pb.QueryRequest) (*pb.QueryReply, error) {
	// Find the transaction in the txs map
	tx, ok := s.txs[req.TxId]
	if !ok {
		return nil, pgx.ErrTxClosed
	}

	// Execute the query
	_, err := tx.Query(ctx, req.SqlStatement, req.Args)
	if err != nil {
		return nil, err
	}
	var results []*pb.Row

	// for rows.Next() {
	// 	columns, err := rows.Values()
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	columnMap := make(map[string]*anypb.Any)
	// 	for i, val := range columns {
	// 		anyVal, err := anypb.
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		columnMap[rows.FieldDescriptions()[i].Name] = anyVal
	// 	}

	// 	results = append(results, &pb.Row{Columns: columnMap})
	// }

	return &pb.QueryReply{Results: results}, nil
}

// Exec


// PrepareTransaction

// CommitPrepared

// Abort
func (s *Server) Abort(ctx context.Context, req *pb.AbortRequest) (*pb.AbortReply, error) {
	// Find the transaction in the txs map
	tx, ok := s.txs[req.TxId]
	if !ok {
		return nil, pgx.ErrTxClosed
	}

	// Rollback the transaction
	err := tx.Rollback(ctx)
	if err != nil {
		return nil, err
	}

	// Remove the transaction from the txs map
	delete(s.txs, req.TxId)

	return &pb.AbortReply{Success: true}, nil
}

func Run(connString string) {
	// Connect to the database
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Create a listener on TCP port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register the server with the gRPC server
	pb.RegisterTransactionServiceServer(grpcServer, &Server{
		pool: pool,
		txs:  make(map[string]pgx.Tx),
	})

	// Register reflection service on gRPC server (optional, used for debugging)
	reflection.Register(grpcServer)

	log.Println("gRPC server is running on port 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
