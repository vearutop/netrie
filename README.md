
# netrie

[![Build Status](https://github.com/vearutop/netrie/workflows/test-unit/badge.svg)](https://github.com/vearutop/netrie/actions?query=branch%3Amaster+workflow%3Atest-unit)
[![Coverage Status](https://codecov.io/gh/vearutop/netrie/branch/master/graph/badge.svg)](https://codecov.io/gh/vearutop/netrie)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/vearutop/netrie)
[![Time Tracker](https://wakatime.com/badge/github/vearutop/netrie.svg)](https://wakatime.com/badge/github/vearutop/netrie)
![Code lines](https://sloc.xyz/github/vearutop/netrie/?category=code)
![Comments](https://sloc.xyz/github/vearutop/netrie/?category=comments)

Netrie is a high-performance, memory-efficient CIDR index library for Go. It provides fast IP lookups for both IPv4 and IPv6 addresses using a trie data structure.

## Features

- Fast IP lookups for both IPv4 and IPv6 addresses
- Memory-efficient trie data structure
- Support for loading from MaxMind GeoIP databases
- Ability to save/load the index to/from files
- Minimization of the trie to reduce memory usage
- Thread-safe lookups
- File-based lookups without loading the full database into memory

## Installation

```bash
go get github.com/vearutop/netrie
```

## Basic Usage

### Creating an Index and Adding Networks

```go
package main

import (
    "fmt"
    "github.com/vearutop/netrie"
)

func main() {
    // Create a new CIDR index
    idx := netrie.NewCIDRIndex()
    
    // Add some networks
    _ = idx.AddCIDR("192.168.1.0/24", "Home Network")
    _ = idx.AddCIDR("10.0.0.0/8", "Private Network")
    _ = idx.AddCIDR("2001:db8::/32", "IPv6 Documentation")
    
    // Look up an IP
    result := idx.Lookup("192.168.1.100")
    fmt.Println("192.168.1.100 belongs to:", result) // Output: 192.168.1.100 belongs to: Home Network
    
    // Look up an IP that doesn't match any network
    result = idx.Lookup("8.8.8.8")
    fmt.Println("8.8.8.8 belongs to:", result) // Output: 8.8.8.8 belongs to: 
    
    // Get statistics
    fmt.Printf("Total networks: %d\n", idx.Len())
    fmt.Printf("Unique names: %d\n", idx.LenNames())
    fmt.Printf("Trie nodes: %d\n", idx.LenNodes())
    
    // Minimize the trie to reduce memory usage
    idx.Minimize()
    fmt.Printf("Trie nodes after minimization: %d\n", idx.LenNodes())
}
```

### Saving and Loading from File

```go
package main

import (
    "fmt"
    "os"
    "github.com/vearutop/netrie"
)

func main() {
    // Create and populate an index
    idx := netrie.NewCIDRIndex()
    _ = idx.AddCIDR("192.168.1.0/24", "Home Network")
    _ = idx.AddCIDR("10.0.0.0/8", "Private Network")
    
    // Add metadata
    idx.Metadata().Description = "My Network Index"
    
    // Save to file
    err := idx.SaveToFile("networks.bin")
    if err != nil {
        panic(err)
    }
    
    // Load from file (loads entire database into memory)
    loadedIdx, err := netrie.LoadFromFile("networks.bin")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Description:", loadedIdx.Metadata().Description)
    fmt.Println("10.1.2.3 belongs to:", loadedIdx.Lookup("10.1.2.3"))
    
    // Open with more options
    f, err := os.Open("networks.bin")
    if err != nil {
        panic(err)
    }
    defer f.Close()
    
    fileIdx, err := netrie.Open(f, func(o *netrie.Options) {
        o.BufferSize = 4096 // Set buffer size for file reading
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Println("10.1.2.3 belongs to:", fileIdx.Lookup("10.1.2.3"))
}
```

### File-Based Lookups Without Loading Full Database

For very large databases or memory-constrained environments, netrie provides the ability to perform lookups directly from the file without loading the entire database into memory:

```go
package main

import (
    "fmt"
    "os"
    "github.com/vearutop/netrie"
)

func main() {
    // Create a file-based lookup index
    // This only loads the header and names into memory, not the entire trie
    idx, err := netrie.OpenFile("large-geoip-database.bin", func(o *netrie.Options) {
        o.BufferSize = 8192 // Adjust buffer size for optimal performance
    })
    if err != nil {
        panic(err)
    }
	defer idx.Close()


	// Perform lookups - nodes will be read from disk as needed
    result := idx.Lookup("81.2.69.145")
    fmt.Println("81.2.69.145 is located in:", result)
    
    // The lookups are thread-safe and can be used concurrently
    // Each lookup will only read the necessary nodes from disk
}
```

Benefits of file-based lookups:
- Significantly lower memory usage as the trie structure remains on disk
- Faster initialization since only metadata and names are loaded initially
- Ability to work with databases that are too large to fit in memory
- Buffered reading for improved performance

### Loading from MaxMind GeoIP Database

```go
package main

import (
    "fmt"
    "github.com/vearutop/netrie"
    "github.com/vearutop/netrie/mmdb"
)

func main() {
    // Create a new CIDR index
    idx := netrie.NewCIDRIndex()
    
    // Load from MaxMind GeoIP City database with city and country information
    err := mmdb.Load(idx, "GeoIP2-City.mmdb", mmdb.CityCountryISOCode)
    if err != nil {
        panic(err)
    }
    
    // Look up an IP
    result := idx.Lookup("81.2.69.145")
    fmt.Println("81.2.69.145 is located in:", result) // Example output: 81.2.69.145 is located in: GB:London
    
    // Minimize the trie to reduce memory usage
    idx.Minimize()
    
    // Save to file for faster loading next time
    err = idx.SaveToFile("geoip-city.bin")
    if err != nil {
        panic(err)
    }
    
    // Create a new index for ASN information
    asnIdx := netrie.NewCIDRIndex()
    asnIdx.Metadata().Description = "ASN Database"
    
    // Load from MaxMind GeoIP ASN database
    err = mmdb.Load(asnIdx, "GeoLite2-ASN.mmdb", mmdb.ASNOrg)
    if err != nil {
        panic(err)
    }
    
    // Look up an IP's ASN information
    asnResult := asnIdx.Lookup("1.0.0.4")
    fmt.Println("1.0.0.4 belongs to:", asnResult) // Example output: 1.0.0.4 belongs to: AS15169 Google Inc.
}
```

## Large Networks Support

For applications that need to handle a large number of networks (more than 2^16), use `NewCIDRLargeIndex()` instead of `NewCIDRIndex()`:

```go
// Create an index that can handle up to 2^32 networks
idx := netrie.NewCIDRLargeIndex()
```

## Thread Safety

The lookup operations are thread-safe and can be used concurrently:

```go
package main

import (
    "fmt"
    "sync"
    "github.com/vearutop/netrie"
)

func main() {
    idx := netrie.NewCIDRIndex()
    _ = idx.AddCIDR("192.168.1.0/24", "Home Network")
    _ = idx.AddCIDR("10.0.0.0/8", "Private Network")
    
    var wg sync.WaitGroup
    wg.Add(100)
    
    for i := 0; i < 100; i++ {
        go func() {
            defer wg.Done()
            result := idx.Lookup("10.1.2.3")
            fmt.Println("10.1.2.3 belongs to:", result)
        }()
    }
    
    wg.Wait()
}
```

## Performance Considerations

- Use `Minimize()` to reduce memory usage after adding all networks
- For frequent lookups, consider saving the index to a file and loading it when needed
- When using file-based lookups, adjust the buffer size based on your access patterns
- For memory-constrained environments, use `Open()` instead of `LoadFromFile()` to avoid loading the entire database into memory

## License

This project is licensed under the MIT License - see the LICENSE file for details.
