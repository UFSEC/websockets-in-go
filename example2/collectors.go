package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"log"
	"io"
	"strings"
	"strconv"
	"encoding/json"
)

type datapoint struct {
	Type string
	Data string
}

type procpoint struct {
	Points []string
}

type proclist struct {
	Type string
	Procs []procpoint
}

func collectTop(data chan []byte) {
	top := exec.Command("top", "-b", "-d", "1")	
	reader, writer := io.Pipe()
	top.Stdout = writer
	scanner := bufio.NewScanner(reader)

	log.Println("Staring top")
	if err := top.Start(); err != nil {
		log.Fatal(err)
	}

	defer top.Wait()

	for scanner.Scan() {
		// log.Println("top scan")
		if strings.Contains(scanner.Text(), "Cpu(s)") {
			// parse the CPU usage out
			tokens := strings.Fields(scanner.Text())
			user, _ := strconv.ParseFloat(tokens[1], 32)
			sys, _ := strconv.ParseFloat(tokens[3], 32)			
			encoded, err := json.Marshal(datapoint{"cpu", fmt.Sprintf("%.1f",user + sys)})
			if err != nil {
				log.Println(err)
				continue
			}
			data <- encoded
		} else if strings.Contains(scanner.Text(), "KiB Mem:") {
			tokens := strings.Fields(scanner.Text())
			used, _ := strconv.ParseFloat(tokens[4], 32)
			total, _ := strconv.ParseFloat(tokens[2], 32)
			percent := (used/total*100)
			encoded, err := json.Marshal(datapoint{"mem", fmt.Sprintf("%.1f", percent)})
			if err != nil {
				log.Println(err)
				continue

			}
			data <- encoded
		} else if strings.Contains(scanner.Text(), "PID") {
			// read the next 10 lines
			var procs []procpoint
			for i := 0; i < 10; i++ {
				scanner.Scan()
				tokens := strings.Fields(scanner.Text())
				procs = append(procs, procpoint{tokens})
			}
			encoded, err := json.Marshal(proclist{"procs", procs})
			if err != nil {
				log.Println(err)
				continue

			}
			data <- encoded
		}
	}
}

func collectIoStat(data chan []byte) {
	iostat := exec.Command("iostat", "-x", "1")	
	reader, writer := io.Pipe()
	iostat.Stdout = writer
	scanner := bufio.NewScanner(reader)

	if err := iostat.Start(); err != nil {
		log.Fatal(err)
	}

	defer iostat.Wait()

	log.Println("Starting to scan")
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "sda") {
			tokens := strings.Fields(scanner.Text())
			util, _ := strconv.ParseFloat(tokens[13], 32)
			encoded, err := json.Marshal(datapoint{"disk", fmt.Sprintf("%.1f", util)})
			if err != nil {
				log.Println(err)
				continue
			}
			data <- encoded
		}
	}
}

