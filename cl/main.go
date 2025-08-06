package main

import "fmt"

func main() {
	input := []string{
		"192.168.0.0/25",
		"192.168.0.128/25",
		"10.0.0.1",
		"10.0.0.2",
		"2001:db8::/126",
		"2001:db8::4/126",
	}

	result, err := ClusterCIDRs(input)
	if err != nil {
		panic(err)
	}

	for _, r := range result {
		fmt.Println(r.String())
	}
}
