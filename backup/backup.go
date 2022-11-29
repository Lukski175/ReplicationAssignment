package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/Lukski175/ReplicationAssignment/time"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedAuctionServiceServer
}

var auctions []*pb.AuctionInfo

func main() {
	auctions = make([]*pb.AuctionInfo, 0)
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create listener tcp on port ownPort
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAuctionServiceServer(grpcServer, &Server{})

	log.Printf("Backup created successfully on port: %d", ownPort)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	for {
		for i := 0; i < len(auctions); i++ {
			auc := auctions[i]
			if auc != nil && !auc.IsFinished {
				//Checks if auction is over
				if time.Now().After(auc.StartTime.AsTime().Add(time.Second * time.Duration(auc.Duration))) {
					log.Printf("Auction for item %s ended", auc.ItemName)
					auc.IsFinished = true
				}
			}
		}
	}
}

func (s *Server) Bid(ctx context.Context, bid *pb.BidMessage) (*pb.AuctionMessage, error) {
	var auction *pb.AuctionInfo
	for i := 0; i < len(auctions); i++ {
		if strings.ToLower(bid.ItemName) == strings.ToLower(auctions[i].ItemName) {
			auction = auctions[i]
			break
		}
	}
	if auction == nil {
		log.Printf("Auction %s not found", bid.ItemName)
		return &pb.AuctionMessage{Message: "Fail"}, nil
	}

	if bid.Amount > auction.MinBid && !auction.IsFinished {
		auction.MinBid = bid.Amount
		auction.ClientName = bid.ClientName
		return &pb.AuctionMessage{Message: "Success"}, nil
	} else {
		return &pb.AuctionMessage{Message: "Fail"}, nil
	}
}

func (s *Server) Result(ctx context.Context, item *pb.AuctionItem) (*pb.AuctionInfo, error) {
	var auction *pb.AuctionInfo
	for i := 0; i < len(auctions); i++ {
		if strings.ToLower(item.ItemName) == strings.ToLower(auctions[i].ItemName) {
			auction = auctions[i]
			break
		}
	}
	if auction == nil {
		return nil, nil
	} else {
		return auction, nil
	}
}

func (s *Server) CreateAuction(ctx context.Context, info *pb.AuctionInfo) (*pb.AuctionMessage, error) {
	info.IsFinished = false
	auctions = append(auctions, info)
	return &pb.AuctionMessage{Message: "Success"}, nil
}

func (s *Server) GetAuctions(e *pb.Empty, stream pb.AuctionService_GetAuctionsServer) error {
	for i := 0; i < len(auctions); i++ {
		stream.Send(auctions[i])
	}
	return nil
}
