package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	pb "github.com/Lukski175/ReplicationAssignment/time"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type backup struct {
	port   int32
	backup pb.AuctionServiceClient
}

var backups []backup
var clientName string = "TestName"

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	numberOfBackups := int(arg1)
	backups = make([]backup, numberOfBackups)

	// For how many backups we want to connect to, loop and connect
	for i := 0; i < numberOfBackups; i++ {
		//Backup 0 has port 5000
		port := int32(5000) + int32(i)

		var conn *grpc.ClientConn
		fmt.Printf("Trying to dial: %v\n", port)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		c := pb.NewAuctionServiceClient(conn)
		backups[i] = backup{port: port, backup: c}
	}
	log.Printf("Connection established")
	log.Printf("")

	// for i := 0; i < len(backups); i++ {
	// 	ack, _ := backups[i].backup.Bid(context.Background(), &pb.BidMessage{})
	// 	log.Printf("Calling bid on backup %d returns: %s", i, ack.Message)
	// }

	for {
		log.Printf("Enter command...")
		log.Printf("Create {Item Name} {Minimum Bid} {Duration (seconds)}")
		log.Printf("Bid {Item Name} {Bid Amount}")
		log.Printf("Result {Item Name}")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.Fields(scanner.Text())

		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()

		switch strings.ToLower(input[0]) {
		case "create":
			Create(input)
			break

		case "bid":
			Bid(input)
			break

		case "result":
			Result(input)
			break
		}

		log.Printf("")
		log.Printf("Press enter to continue")
		fmt.Scanln()

		cmd = exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func Create(input []string) {
	bid, _ := strconv.Atoi(input[2])
	duration, _ := strconv.Atoi(input[3])
	acks := make([]*pb.AuctionMessage, 0)
	for i := 0; i < len(backups); i++ {
		ack, err := backups[i].backup.CreateAuction(context.Background(), &pb.AuctionInfo{ItemName: input[1], MinBid: int32(bid), Duration: int32(duration), StartTime: timestamppb.Now()})
		if err != nil {
			BackupCrashed(backups[i])
			i--
		} else {
			acks = append(acks, ack)
		}
	}
	if len(acks) > 0 {
		log.Printf("Received message: %s", acks[0].Message)
	}
}

func Bid(input []string) {
	amount, _ := strconv.Atoi(input[2])
	acks := make([]*pb.AuctionMessage, 0)
	for i := 0; i < len(backups); i++ {
		ack, err := backups[i].backup.Bid(context.Background(), &pb.BidMessage{ItemName: input[1], Amount: int32(amount), ClientName: clientName})
		if err != nil {
			BackupCrashed(backups[i])
			i--
		} else {
			acks = append(acks, ack)
		}
	}
	if len(acks) > 0 {
		log.Printf("Received message: %s", acks[0].Message)
	}
}

func Result(input []string) {
	acks := make([]*pb.AuctionInfo, 0)
	for i := 0; i < len(backups); i++ {
		ack, err := backups[i].backup.Result(context.Background(), &pb.AuctionItem{ItemName: input[1]})
		if err != nil {
			BackupCrashed(backups[i])
			i--
		} else {
			acks = append(acks, ack)
		}
	}
	if len(acks) > 0 {
		PrintAuction(acks[0])
	}
}

func PrintAuction(ack *pb.AuctionInfo) {
	log.Printf("Auction for item %s\nCurrent Bid: %d\nAuction Finished: %v", ack.ItemName, ack.MinBid, ack.IsFinished)
}

func BackupCrashed(back backup) {
	log.Printf("Backup with port %d crashed", back.port)
	newBackups := make([]backup, 0)
	for _, item := range backups {
		if item != back {
			newBackups = append(newBackups, item)
		}
	}
	backups = newBackups
}
