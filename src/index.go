package main

import (
	"fmt"
	"github.com/sandertv/go-raknet"
	"sync"
	"time"
)

func main() {
	var ip string
	startPort := 1
	endPort := 65535
	workerCount := 1000

	for {
		fmt.Println("\nPort Scanner Menu")
		fmt.Println("1. Set IP address")
		fmt.Println("2. Set port range")
		fmt.Println("3. Set worker count")
		fmt.Println("4. Start scanning")
		fmt.Println("5. Exit")
		fmt.Print("Choose an option: ")

		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil {
			fmt.Println("Invalid input, please try again.")
			continue
		}

		switch choice {
		case 1:
			fmt.Print("Enter IP address: ")
			_, err := fmt.Scanln(&ip)
			if err != nil {
				fmt.Println("Invalid input, please try again.")
			}
		case 2:
			fmt.Print("Enter start port: ")
			_, err := fmt.Scanln(&startPort)
			if err != nil {
				fmt.Println("Invalid input, please try again.")
			}
			fmt.Print("Enter end port: ")
			_, err = fmt.Scanln(&endPort)
			if err != nil {
				fmt.Println("Invalid input, please try again.")
			}
		case 3:
			fmt.Print("Enter worker count: ")
			_, err := fmt.Scanln(&workerCount)
			if err != nil {
				fmt.Println("Invalid input, please try again.")
			}
		case 4:
			if ip == "" || startPort == 0 || endPort == 0 || workerCount == 0 {
				fmt.Println("Please set all parameters before starting the scan.")
				continue
			}
			startScanning(ip, startPort, endPort, workerCount)
		case 5:
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice, please try again.")
		}
	}
}

func startScanning(ip string, startPort, endPort, workerCount int) {
	totalPorts := endPort - startPort + 1
	ports := make(chan int, 1000)
	results := make(chan int, totalPorts)
	progress := make(chan int, totalPorts)
	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ip, ports, results, progress, &wg)
	}

	go func() {
		for port := startPort; port <= endPort; port++ {
			ports <- port
		}
		close(ports)
	}()

	go func() {
		wg.Wait()
		close(results)
		close(progress)
	}()

	go func() {
		scannedPorts := 0
		for range progress {
			scannedPorts++
			fmt.Printf("\rProgress: %d/%d ports scanned", scannedPorts, totalPorts)
		}
		fmt.Println()
	}()

	var openPorts []int
	for port := range results {
		if port != 0 {
			openPorts = append(openPorts, port)
		}
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	fmt.Printf("Scanning completed in %v\n", duration)

	if len(openPorts) > 0 {
		fmt.Println("Open ports:")
		for _, port := range openPorts {
			fmt.Printf("%d\n", port)
		}
	} else {
		fmt.Println("No open ports found.")
	}
}

func worker(ip string, ports <-chan int, results chan<- int, progress chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for port := range ports {
		scanPort(ip, port, results)
		progress <- 1
	}
}

func scanPort(ip string, port int, results chan<- int) {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := raknet.DialTimeout(address, 500*time.Millisecond)
	if err == nil {
		results <- port
		_ = conn.Close()
	} else {
		results <- 0
	}
}
