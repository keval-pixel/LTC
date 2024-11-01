package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type Message struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

func main() {
	// Connect to the server
	reader := bufio.NewReader(os.Stdin)
	var Svc string
	toggle := make(chan []byte)
	buffer := make([]byte, 1)
	running := true
	var input string
	var IpAddr string
	var err error

	welcome_message()
	for running {
		for IpAddr == "" {
			fmt.Print("Enter the IP address/FQDN of the LTC machine: ")
			IpAddr, err = reader.ReadString('\n')
			IpAddr = strings.TrimSpace(IpAddr)
		}
		if IpAddr == "exit" {
			break
		}
		Svc = IpAddr + ":8080"
		IpAddr = ""
		if err != nil {
			color.Red("Error reading input:", err)
			continue
		}
		conn, err := net.Dial("tcp", Svc)
		if err != nil {
			color.Red("Error connecting: %v\n", err)
			continue
		}

		color.Green("Agent Connected!")
		color.Yellow("Waiting for the load testing to start...")

		sending := false

		go func() {
			for conn != nil {
				_, err = conn.Read(buffer)
				if err == io.EOF {
					fmt.Println("Connection closed by LTC")
					conn.Close()
					conn = nil
					return
				}
				if buffer[0] == 0x11 {
					//fmt.Println("Received byte from test server!")
					toggle <- buffer
				}
			}

		}()

		done := make(chan struct{})

		go func() {
		sending_loop:
			for {
				select {
				case <-done:
					return
				case <-toggle:
					sending = !sending
					if sending {
						color.Green("Testing started... sending metrics.\n")
					} else {
						color.Red("Testing stopped.\n")
					}
				default:
					if conn == nil {
						break sending_loop
					}
					if sending {
						//Get CPU and Memory usage from the machine
						cpuUsage, err := cpu.Percent(time.Second, false)
						if err != nil {
							color.Red("Error getting CPU usage\n", err)
						}
						memInfo, err := mem.VirtualMemory()
						if err != nil {
							color.Red("Error getting memory usage:", err)
						}

						message := Message{
							CPU:    cpuUsage[0],
							Memory: float64(memInfo.Used) / float64(memInfo.Total) * 100,
						}

						// Marshal the Message struct into JSON
						jsonData, err := json.Marshal(message)
						if err != nil {
							color.Red("Error marshaling JSON:", err)
							sending = false
							break sending_loop
						}

						fmt.Printf("CPU: %.1f%%, memory used: %.1f%%\n", message.CPU, message.Memory)

						// Send the JSON data to the server
						_, err = conn.Write(jsonData)
						if err != nil {
							color.Red("Error sending: %s", err.Error())
							sending = false
							break sending_loop
						}
					}
				}
			}
			running = false
			fmt.Println("Do you want to reconnect to the LTC tool?[y/N]")

			return

		}()

		for {
			input, _ = reader.ReadString('\n')
			input = strings.ToLower(input)
			input = strings.TrimSpace(input)

			if input == "exit" || input == "2" {
				color.Cyan("Exiting...")
				close(done)
				running = false
				break
			} else if input == "y" && running == false {
				running = true
				break
			} else if input == "n" || input == "" && running == false {
				color.Cyan("Exiting...")
				close(done)
				running = false
				break
			} else {
				continue
			}
		}
	}
	color.Green("Program exited.")
}

func welcome_message() {
	color.HiMagenta("Welcome to LTC (Load Testing Controller) Metrics Agent\n")
	println("Please select on of the following options:\n1.Connect to the LTC\n2.Exit\n")
}

func IntToBytes(num int) []byte {
	buf := new(bytes.Buffer)
	// Write the int into the buffer as a byte array
	err := binary.Write(buf, binary.LittleEndian, int32(num))
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}
