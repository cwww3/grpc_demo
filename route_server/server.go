package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	pb "github.com/cwww3/grpc_demo/route"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedRouteGuideServer
	features []*pb.Feature
}

func (s *Server) GetFeature(ctx context.Context, p *pb.Point) (*pb.Feature, error) {
	for _, feature := range s.features {
		if proto.Equal(feature.Location, p) {
			return feature, nil
		}
	}
	return nil, errors.New("not found")
}
func (s *Server) ListFeatures(r *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
	for _, feature := range s.features {
		if feature.Location.X >= r.Hi.X && feature.Location.Y <= r.Hi.Y &&
			feature.Location.X <= r.Lo.X && feature.Location.Y >= r.Lo.Y {
			if err := stream.Send(feature); err != nil {
				return err
			}
			time.Sleep(time.Second)
		}
	}
	return nil
}
func (s *Server) RecordRoute(stream pb.RouteGuide_RecordRouteServer) error {
	startTime := time.Now()
	var point_cnt, distance int32
	var prevPoint *pb.Point

	for true {
		point, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.RouteSummary{
				PointCount:  point_cnt,
				Distance:    distance,
				ElapsedTime: int32(time.Now().Sub(startTime).Seconds()),
			})
		}
		if err != nil {
			return err
		}
		if point != nil {
			distance += calcDistance(point, prevPoint)
		}
		point_cnt++
		prevPoint = point
	}
	return nil
}
func (s *Server) Recommend(stream pb.RouteGuide_RecommendServer) error {
	var n int
	for true {
		request, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("receive eof break")
			break
		}
		if err != nil {
			fmt.Println("receive err", err)
			return err
		}
		feature, err := s.RecommendOnce(request)
		if err != nil {
			return err
		}
		err = stream.Send(feature)
		if err == io.EOF {
			fmt.Println("send eof break")
			break
		}
		if err != nil {
			fmt.Println("send err", err)
			return err
		}
		n++
		fmt.Println("Request :", n)
	}
	return nil
}

// RecommendOnce ????????????????????????????????????
func (s *Server) RecommendOnce(request *pb.RecommendationRequest) (*pb.Feature, error) {
	if request.Mode == pb.RecommendationMode_GetFarthest {
		return &pb.Feature{
			Name: "farthest",
			Location: &pb.Point{
				X: 100,
				Y: 100,
			},
		}, nil
	} else if request.Mode == pb.RecommendationMode_GetNearest {
		return &pb.Feature{
			Name: "nearest",
			Location: &pb.Point{
				X: 0,
				Y: 0,
			},
		}, nil
	} else {
		return nil, errors.New("mode is undefined")
	}
}

func NewServer() *Server {
	return &Server{
		features: []*pb.Feature{
			{
				Name: "A???",
				Location: &pb.Point{
					X: 1,
					Y: 1,
				},
			},
			{
				Name: "B???",
				Location: &pb.Point{
					X: 5,
					Y: 5,
				},
			},
			{
				Name: "C???",
				Location: &pb.Point{
					X: 10,
					Y: 10,
				},
			},
		},
	}
}

func main() {
	// GRPC ??????
	grpcServer := grpc.NewServer()
	pb.RegisterRouteGuideServer(grpcServer, NewServer())
	lis, err := net.Listen("tcp", ":7001")
	if err != nil {
		log.Fatalln("cannot create a listener")
	}
	defer lis.Close()
	go func() {
		log.Fatalln(grpcServer.Serve(lis))
	}()

	// gRPC-Gateway ??????
	conn, err := grpc.DialContext(
		context.Background(),
		"127.0.0.1:7001",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = pb.RegisterRouteGuideHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    ":7002",
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:7002")
	log.Fatalln(gwServer.ListenAndServe())
}

// TODO ??????????????????????????? ???????????????
func calcDistance(p1, p2 *pb.Point) int32 {
	return 10
}
