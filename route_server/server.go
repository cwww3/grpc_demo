package main

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net"
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

// RecommendOnce 获得最近或最远的点的信息
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
				Name: "A点",
				Location: &pb.Point{
					X: 1,
					Y: 1,
				},
			},
			{
				Name: "B点",
				Location: &pb.Point{
					X: 5,
					Y: 5,
				},
			},
			{
				Name: "C点",
				Location: &pb.Point{
					X: 10,
					Y: 10,
				},
			},
		},
	}
}

func main() {
	grpcServer := grpc.NewServer()
	pb.RegisterRouteGuideServer(grpcServer, NewServer())
	lis, err := net.Listen("tcp", "localhost:5000")
	if err != nil {
		log.Fatalln("cannot create a listener")
	}
	defer lis.Close()
	log.Fatalln(grpcServer.Serve(lis))
}

// TODO 计算两点之间的距离 返回非负数
func calcDistance(p1, p2 *pb.Point) int32 {
	return 10
}
