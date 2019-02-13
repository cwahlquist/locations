package service

import (
    "encoding/json"
    "errors"
    "golang.org/x/net/context"
    "github.com/davecgh/go-spew/spew"
    "io/ioutil"
    "log"
	pb "locations/api/go"
)

var locationList []*pb.LocationType

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) RegisterLocations(ctx context.Context, r *pb.RegisterLocationRequest) (*pb.RegisterLocationResponse, error) {
    err := registerLocations(r)
    return &pb.RegisterLocationResponse{}, err
}

func (s *Service) GetLocations(ctx context.Context, r *pb.LocationsRequest) (*pb.LocationsResponse, error) {
    var results []*pb.LocationType
    results = filterLocations(r, results)
    return &pb.LocationsResponse{Locations: results}, nil
}

func (s *Service) GetLocation(ctx context.Context, r *pb.LocationRequest) (*pb.LocationResponse, error) {
    ll, err := findLocation(r)
    return &pb.LocationResponse{Location: ll}, err
}

func ReadLocations(subFile string) error {
    err := loadLocations(subFile,&locationList)
    spew.Dump("init", locationList)
    return err
}

func loadLocations(subFile string, locations *[]*pb.LocationType) error {
    subs_json, err := readSubscriptions(subFile)
    if err != nil {
        return err
    }
    err = json.Unmarshal(subs_json, locations)
    if err != nil {
        return err
    }
    return nil
}

func filterLocations(r *pb.LocationsRequest, results []*pb.LocationType) []*pb.LocationType {
    for _, ll := range locationList {
        match := true
        // cloud is required
        if (ll.Cloud != r.Cloud) {
            match = false
        }
        // region is optional
        if match && (len(r.Region) > 0) && (ll.Region != r.Region) {
            match = false
        }
        // type is optional
        if match && (r.Type != 0) && (ll.Type != r.Type) {
            match = false
        }
        if match {
            results = append(results, ll)
        }
    }
    return results
}

func findLocation(r *pb.LocationRequest) (*pb.LocationType,error) {
    for _, ll := range locationList {
        log.Printf("compare: %s:%s %s:%s", ll.SubId, r.SubId, ll.Cloud, r.Cloud)
        if ( ll.SubId == r.SubId ) && ( ll.Cloud == r.Cloud) {
            return ll, nil
        }
    }
    return nil, errors.New("Error: No matching location  found.")
}

func readSubscriptions(subFile string) ([]byte,error) {
    nbytes, err := ioutil.ReadFile(subFile)
    if err != nil {
        return nil, err
    }
    return nbytes, nil
}

func registerLocations(r *pb.RegisterLocationRequest) (error) {
    for _, ll := range r.GetLocations() {
        locationList = append(locationList,ll)
    }
    return nil
}


