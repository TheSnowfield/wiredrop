package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	"wiredrop/wiredrop"
)

type Configuration struct {
	Listen string `json:"listen"`
	SSL    *struct {
		Cert string `json:"cert"`
		Key  string `json:"key"`
	} `json:"ssl"`
	Peer *struct {
		Timeout    *uint   `json:"timeout"`
		BufferSize *uint64 `json:"buffersize"`
	} `json:"peer"`
}

func main() {

	// parse the cmdline
	cmds, err := commandParser(os.Args[1:])
	if err != nil {
		println(err)
		os.Exit(1)
	}

	// print usage text
	if len(cmds) == 0 || existKey(cmds, "--help") {
		printHelp()
		os.Exit(0)
	}

	// parse the cmd line
	if configPath, exist := existKeyWithValue(cmds, "-C", "--config"); exist {

		// parse the configuration file
		config, err := parseConfigurationJson(*configPath)
		if err != nil {
			println("invalid configuration file given")
			os.Exit(1)
		}

		// set peer max wait timeout
		if config.Peer != nil && config.Peer.Timeout != nil {
			wiredrop.SetPeerTimeout(time.Duration(*config.Peer.Timeout) * time.Second)
			println(fmt.Sprintf("set peer wait timeout %d", config.Peer.Timeout))
		}

		// set peer initial buffer size
		if config.Peer != nil && config.Peer.BufferSize != nil {
			wiredrop.SetPeerInitialBufferSize(*config.Peer.BufferSize)
			println(fmt.Sprintf("set buffer size to %d", config.Peer.BufferSize))
		}

		log.Print("wiredrop is listening on " + config.Listen)

		// start the server
		if config.SSL != nil {
			err = wiredrop.StartTLS(config.Listen, config.SSL.Cert, config.SSL.Key)
		} else {
			err = wiredrop.Start(config.Listen)
		}
	}

	if err != nil {
		println(err)
		os.Exit(1)
	}
}

// command parser
func commandParser(cmdline []string) (map[string]string, error) {

	// print usage
	if len(cmdline) == 0 {
		printHelp()
		os.Exit(0)
	}

	// search the command line arguments
	var cmds = make(map[string]string)
	for i, part := range cmdline {

		switch part {
		case "-C", "--config":
			if i+1 >= len(cmdline) {
				println(fmt.Sprintf("%s: please specify the path of an valid configuration file", part))
				os.Exit(1)
			}

			cmds[part] = cmdline[i+1]
			break

		case "-H", "--help":
			cmds[part] = ""
			break
		}
	}

	return cmds, nil
}

func parseConfigurationJson(configPath string) (*Configuration, error) {

	// read file
	file, err := os.ReadFile(configPath)
	if err != nil {
		println(fmt.Printf("failed to read configuration file: %s", configPath))
		return nil, err
	}

	// parse the json
	config := new(Configuration)
	err = json.Unmarshal(file, &config)
	if err != nil {
		println(fmt.Printf("failed to parse configuration file"))
		return nil, err
	}

	return config, nil
}

func existKey(kv map[string]string, k string) bool {
	if _, exist := kv[k]; exist {
		return true
	}
	return false
}

func existKeyWithValue(kv map[string]string, k ...string) (*string, bool) {

	for _, part := range k {
		if val, exist := kv[part]; exist {
			return &val, true
		}
	}

	return nil, false
}

func printHelp() {
	println("Usage: wiredrop -C <config.json>")
	println("")
	println("-C, --config \t specify the configuration file")
	println("-H, --help \t prints this help")
	println("")
	println("Examples:")
	println("wiredrop --config ./config.json Starts the wiredrop server with config.json as initial configuration")
	println("")
	println("wiredrop (c) TheSnowfield\nLICENSED under MIT")
}
