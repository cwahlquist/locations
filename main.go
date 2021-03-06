package main

import (
    "flag"
    "fmt"
    "log"
    "time"
    "encoding/json"
    "net"
    "net/http"
    "path/filepath"
    locs "ares.locations"
    s "locations/service"
    "google.golang.org/grpc"
    "github.com/gin-gonic/gin"
    // needed for postman proxy
    _ "github.com/jnewmano/grpc-json-proxy/codec"
)

var (
    port = flag.Int("port", 31400, "The server grpc port")
    rest = flag.Int("http", 8080, "The server http port")
)


func main() {

    subFile := filepath.Join("/etc/config", "locations.json")
    log.Printf("%s",subFile);

    err := s.ReadLocations(subFile)
    if err != nil {
        log.Printf("Failed to load locations: %s", err)
        subFile = filepath.Join(".", "locations.json")
        err = s.ReadLocations(subFile)
        if err != nil {
            log.Fatalf("Failed to load locations: %s", err)
        }
    }

    // get env vars
    flag.Parse()

    // start listening tcp:host:port
    listen, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
    if err != nil {
        log.Fatalf("Failed to listen: %s", err)
    }

    // inject dependencies

    // initialize service layer
    srv := s.NewService()

    // create grpc server and apply middleware
    grpcServer := grpc.NewServer()

    // register missions PB with grpcServer
    locs.RegisterLocationsServiceServer(grpcServer, srv)

    router := gin.Default()

    s := &http.Server{
        Addr:           fmt.Sprintf(":%d", *rest),
        Handler:        router,
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    router.GET("/locations/:cloud/:region", func(c *gin.Context) {
        // Parse parameters
        log.Println("/locations/%s/%s", c.Param("cloud"), c.Param("region"))
        // Call locations service
        req := &locs.LocationsRequest{}
        req.Cloud = c.Param("cloud")
        req.Region = c.Param("region")
        req.SubId = c.Query("subid")
        ltype := c.Query("type")
        if len(ltype) > 0 {
            fmt.Sscanf(ltype, "%d", &req.Type)
        }
        results, err := getLocations(c, req, srv)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        } else {
            c.JSON(http.StatusOK, gin.H{"result": results})
        }
    })
    router.GET("/locations/:cloud", func(c *gin.Context) {
        // Parse parameters
        log.Println("/locations/%s", c.Param("cloud"))
        // Call locations service
        req := &locs.LocationsRequest{}
        req.Cloud = c.Param("cloud")
        results, err := getLocations(c, req, srv)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        } else {
            c.JSON(http.StatusOK, gin.H{"result": results})
        }
    })
    router.GET("/location/:cloud/:region/:subid", func(c *gin.Context) {
        // Parse parameters
        log.Println("/location/%s/%s", c.Param("cloud"), c.Param("subid"))
        // Call locations service
        req := &locs.LocationRequest{}
        req.Cloud = c.Param("cloud")
        req.Region = c.Param("region")
        req.SubId = c.Param("subid")
        results, err := getLocation(c, req, srv)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        } else {
            c.JSON(http.StatusOK, gin.H{"result": results})
        }
    })
    router.GET("/health", func(c *gin.Context) {
        results := make(map[string]string)
        results["status"] = "Healthy!"
        c.JSON(http.StatusOK, gin.H{"result": results})
    })
    router.GET("/readiness", func(c *gin.Context) {
        results := make(map[string]string)
        results["status"] = "Ready!"
        c.JSON(http.StatusOK, gin.H{"result": results})
    })
    router.GET("/", func(c *gin.Context) {
        results := make(map[string]string)
        results["status"] = "Healthy!"
        c.JSON(http.StatusOK, gin.H{"result": results})
    })

    go func() {
        log.Printf("Locations rest service started on 0.0.0.0:%d", *rest)
        s.ListenAndServe()
    }()
    log.Printf("Locations service started on 0.0.0.0:%d", *port)

    // start gRPC server
    err = grpcServer.Serve(listen)
    if err != nil {
        log.Fatalf("gRpc Server failed to start")
    }
}

func getLocations(c *gin.Context, req *locs.LocationsRequest, srv *s.Service) ([]map[string]interface{},error) {
    res, err := srv.GetLocations(c, req)
    if err == nil {
        log.Printf("%s",res)
        results := make([]map[string]interface{},len(res.GetLocations()))
        for ii, location := range res.GetLocations() {
            result, err := locationToMap(location)
            if err != nil {
                return nil, err
            }
            results[ii] = result
        }
        return results, nil
    }
    return nil, nil
}

func getLocation(c *gin.Context, req *locs.LocationRequest, srv *s.Service) (map[string]interface{},error) {
    if res, err := srv.GetLocation(c, req); err == nil {
        result, err := locationToMap(res.GetLocation())
        return result, err
    } else {
        return nil, err
    }
}

func locationToMap(location *locs.LocationType) (map[string]interface{},error) {
    locations, err := json.Marshal(location)
    if err != nil {
        return nil, err
    }
    var nmap map[string]interface{}
    err = json.Unmarshal(locations, &nmap)
    if err != nil {
        return nil, err
    }
    return nmap, nil
}

