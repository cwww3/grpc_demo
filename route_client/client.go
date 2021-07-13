package main

import (
	"bufio"
	"context"
	"fmt"
	pb "github.com/cwww3/grpc_demo/route"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	// 1.忽略证书
	// 2.阻塞知道拨号成功
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

	go func() {
		feature, err := stream.Recv()
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println("Recommend: ", feature)
	}()

	reader := bufio.NewReader(os.Stdin)

	for true {
		var mode int32
		var request = pb.RecommendationRequest{
			Point: new(pb.Point),
		}
		fmt.Print("Enter Mode (0 for farthest 1 for nearest) : ")
		readIntFromCommandLine(reader, &mode)
		request.Mode = pb.RecommendationMode(mode)
		fmt.Print("Enter X: ")
		readIntFromCommandLine(reader, &request.Point.X)
		fmt.Print("Enter Y: ")
		readIntFromCommandLine(reader, &request.Point.Y)

		if err := stream.Send(&request); err != nil {
			log.Fatalln(err)
		}
		time.Sleep(time.Millisecond * 1000)
	}
}

func readIntFromCommandLine(reader *bufio.Reader, num *int32) {
	_, err := fmt.Fscanf(reader, "%d\n", num)
	if err != nil {
		log.Fatalln(err)
	}
}
