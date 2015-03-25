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
	"math/rand"
	"time"
	"runtime"
)

var randProcs = []procpoint {
	procpoint{[]string{"4785","baker","20","0","1292676","283608","37388","S","3.0","7.2","22:27.70","atom"}},
	procpoint{[]string{"2185","baker","20","0","1735464","282564","29208","S","2.0","7.2","58:50.64","cinnamon"}},
	procpoint{[]string{"2505","baker","20","0","1560780","204308","46176","S","2.0","5.2","17:48.52","chrome"}},
	procpoint{[]string{"23320","baker","20","0","758232","38876","19000","S","2.0","1.0","0:00.56","chrome"}},
	procpoint{[]string{"7","root","20","0","0","0","0","S","1.0","0.0","0:15.94","rcu_sched"}},
	procpoint{[]string{"1492","root","20","0","428964","88436","71276","S","1.0","2.2","20:26.72","Xorg"}},
	procpoint{[]string{"2465","baker","20","0","1755096","188552","50052","S","1.0","4.8","30:57.38","chrome"}},
	procpoint{[]string{"23313","baker","20","0","120356","3508","2660","S","1.0","0.1","0:00.05","example2"}},
	procpoint{[]string{"1","root","20","0","33852","2688","1236","S","0.0","0.1","0:03.47","init"}},
}

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

func generateRandProcData(data chan []byte) {
	for {
		var procs []procpoint
		randIndices := rand.Perm(len(randProcs))
		for _, index := range randIndices {
		    procs = append(procs, randProcs[index])
		}
		encoded, _ := json.Marshal(proclist{"procs", procs})
		data <- encoded
		encoded, _ = json.Marshal(datapoint{"cpu", fmt.Sprintf("%.1f", rand.Float32() * 100)})
		data <- encoded
		encoded, _ = json.Marshal(datapoint{"mem", fmt.Sprintf("%.1f", rand.Float32() * 100)})
		data <- encoded
		time.Sleep(1 * time.Second)
	}
}

func generateRandDiskData(data chan []byte) {
	for {
		encoded, err := json.Marshal(datapoint{"disk", fmt.Sprintf("%.1f", rand.Float32() * 100)})
		if err != nil {
			log.Println(err)
			continue
		}
		data <- encoded
		time.Sleep(1 * time.Second)
	}
}

func collectTop(data chan []byte) {
	if runtime.GOOS == "windows" {
    	generateRandProcData(data)
    	return
	}

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
	if runtime.GOOS == "windows" {
    	generateRandDiskData(data)
    	return
	}

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

