# mdnsresolver

Grpc MDNS Resolver

- A Grpc name resolver by using zeroconf mdns.
- Useful when developing microservices locally for service discovery.
- It comes with a small ~250 LOC mdns client to find service endpoints. Therefore it won't bloat your binaries.

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/adityak368/mdnsresolver) [![Go Report Card](https://goreportcard.com/badge/github.com/adityak368/mdnsresolver)](https://goreportcard.com/report/github.com/adityak368/mdnsresolver)

### USAGE

```go
// Import the module
import (
    "github.com/adityak368/mdnsresolver"
    "google.golang.org/grpc"
)

// if schema is 'mdns' then grpc will use mdnsresolver to resolve addresses
cc, err := grpc.Dial("mdns://service/serviceInstanceName.domain", grpc.WithResolvers(mdnsresolver.NewBuilder()))
// Ex: mdns://MyApp/MyAwesomeMicroService.local
```

More information on naming can be found in [grpc naming docs](https://github.com/grpc/grpc/blob/master/doc/naming.md)

### Client Side Load Balancing

You need to pass grpc.WithBalancerName option to grpc on dial:

```go
grpc.DialContext(ctx,  "mdns://service/serviceInstanceName.domain", grpc.WithResolvers(mdnsresolver.NewBuilder()), grpc.WithBalancerName("round_robin"), grpc.WithInsecure())
```

This will create subconnections for each available service endpoints.

### MDNS Service setup

To setup a service for the client to discover via mdns, have a look at the examples in

https://github.com/grandcat/zeroconf

https://github.com/hashicorp/mdns

or if you are using [Ego Framework](https://github.com/adityak368/ego), you can just setup a registry using:

```go

import (
	"github.com/adityak368/ego/registry"
	"github.com/adityak368/ego/registry/mdns"
)

service := "MyApp"
serviceName := "MyAwesomeMicroService"
domain := "local"

reg := mdns.New(service, domain)
reg.Init(registry.Options{})

reg.Register(registry.Entry{
    Name:   serviceName,
    Address: "localhost:1212",
    Version: "1.0.0",
})
err := reg.Watch()
if err != nil {
    return err
}
defer reg.Deregister(serviceName)
defer reg.CancelWatch()

```

Using [Ego Server](https://github.com/adityak368/ego/server):

```go

import (
    "github.com/adityak368/ego/registry"
    "github.com/adityak368/ego/registry/mdns"
    "github.com/adityak368/ego/server"
    grpcServer "github.com/adityak368/ego/server/grpc"
    "google.golang.org/grpc"
)

service := "MyApp"
serviceName := "MyAwesomeMicroService"
domain := "local"

reg := mdns.New(service, domain)
reg.Init(registry.Options{})

srv := grpcServer.New()

err := srv.Init(server.Options{
    Name: serviceName,
    Host: "localhost",
    Registry: reg,
    Version:  "1.0.0",
})
if err != nil {
    log.Fatal(err)
}

grpcHandle, ok := srv.Handle().(*grpc.Server)
if !ok {
    log.Fatal("Invalid Grpc Server Handle")
}

// Register any protobuf services here

if err := srv.Run(); err != nil {
    log.Fatal(err)
}

```
