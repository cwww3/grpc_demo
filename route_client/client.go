package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	pb "github.com/cwww3/grpc_demo/route"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	// 1.忽略证书
	// 2.阻塞直到拨号成功
	conn, err := grpc.Dial("localhost:5000", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	client := pb.NewRouteGuideClient(conn)
	runForth(client)
}

func runFirst(client pb.RouteGuideClient) {
	feature, err := client.GetFeature(context.Background(), &pb.Point{
		X: 5,
		Y: 5,
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(feature)
}

func runSecond(client pb.RouteGuideClient) {
	serverStream, err := client.ListFeatures(context.Background(), &pb.Rectangle{
		Hi: &pb.Point{X: 0, Y: 10},
		Lo: &pb.Point{X: 10, Y: 0},
	})
	if err != nil {
		log.Fatalln(err)
	}
	for {
		feature, err := serverStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(feature)
	}
}

func runThird(client pb.RouteGuideClient) {
	// dummy data
	points := []*pb.Point{
		{X: 0, Y: 0},
		{X: 2, Y: 2},
		{X: 9, Y: 9},
	}

	clientStream, err := client.RecordRoute(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, point := range points {
		if err = clientStream.Send(point); err != nil {
			log.Fatalln(err)
		}
		time.Sleep(time.Second)
	}
	summary, err := clientStream.CloseAndRecv()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(summary)
}

func runForth(client pb.RouteGuideClient) {
	stream, err := client.Recommend(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err = stream.CloseSend(); err != nil {
			log.Fatalln("close err ", err)
		}
	}()

	// receive from server
	go func() {
		for true {
			feature, err := stream.Recv()
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Recommend: ", feature)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)

	var n int

	// send to server
	for true {
		var request = pb.RecommendationRequest{
			Point: new(pb.Point),
		}
		console(scanner, &request)
		if err := stream.Send(&request); err != nil {
			log.Fatalln(err)
		}
		time.Sleep(time.Millisecond * 1000)

		n++
		// test client close
		if n == 5 {
			break
		}
	}
}

func readIntFromCommandLine(scanner *bufio.Scanner, num *int32) error {
	// block for user input
	if scanner.Scan() {
		text := scanner.Text()
		i, err := strconv.ParseInt(text, 10, 32)
		if err != nil {
			return err
		}
		*num = int32(i)
		return nil
	} else {
		return errors.New("scan nil")
	}
}

func console(scanner *bufio.Scanner, request *pb.RecommendationRequest) {
	var (
		mode       int32
		f1, f2, f3 bool
		err        error
	)

	for !f1 {
		fmt.Print("Enter Mode (0 for farthest 1 for nearest) : ")
		if err = readIntFromCommandLine(scanner, &mode); err != nil {
			fmt.Println(err)
			continue
		}
		f1 = true
		request.Mode = pb.RecommendationMode(mode)
	}

	for !f2 {
		fmt.Print("Enter X: ")
		if err = readIntFromCommandLine(scanner, &request.Point.X); err != nil {
			fmt.Println(err)
			continue
		}
		f2 = true
	}

	for !f3 {
		fmt.Print("Enter Y: ")
		if err = readIntFromCommandLine(scanner, &request.Point.Y); err != nil {
			fmt.Println(err)
			continue
		}
		f3 = true
	}
}
