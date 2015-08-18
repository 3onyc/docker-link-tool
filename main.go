package main

import (
	"flag"
	"fmt"
	docker "github.com/fsouza/go-dockerclient"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	DockerUrl  *string = flag.String("docker", "unix:///var/run/docker.sock", "Docker URL")
	ClientIp   *string = flag.String("cip", "", "Client IP (Optional)")
	ServerIp   *string = flag.String("sip", "", "Server IP (Optional)")
	ClientName *string = flag.String("cname", "", "Client Name (Optional)")
	ServerName *string = flag.String("sname", "", "Server Name (Optional)")
	Port       *uint   = flag.Uint("port", 0, "Server port")
	Action     *string = flag.String("action", "", "delete/add")

	Docker *docker.Client
)

func main() {
	flag.Parse()

	if *ClientIp == "" && *ClientName == "" {
		fmt.Println("Erorr: need one of cname or cip")
		flag.Usage()
		return
	}
	if *ServerIp == "" && *ServerName == "" {
		fmt.Println("Error: need one of sname or sip")
		flag.Usage()
		return
	}

	if *ClientName != "" && *ClientIp == "" {
		var err error
		*ClientIp, err = resolveDockerIp(*ClientName)

		if err != nil {
			fmt.Println("ERR", err)
			return
		}
	}
	if *ServerName != "" && *ServerIp == "" {
		var err error
		*ServerIp, err = resolveDockerIp(*ServerName)

		if err != nil {
			fmt.Println("ERR", err)
			return
		}
	}

	// Only use IPs
	if err := iptablesLink(*ClientIp, *ServerIp, *Port, *Action); err != nil {
		fmt.Println("ERR", err)
		return
	}
}

func iptablesLink(cip, sip string, p uint, a string) error {
	var f string
	if a == "add" {
		f = "I"
	} else if a == "delete" {
		f = "D"
	}

	fmt.Printf("Linking client %s to server %s on port %d\n", cip, sip, p)

	var (
		out []byte
		err error
	)

	out, err = exec.Command(
		"/usr/sbin/iptables",
		"-w", // Wait for xlock
		"-"+f, "FORWARD",
		"-s", cip,
		"-d", sip,
		"-p", "tcp",
		"--dport", strconv.FormatUint(uint64(p), 10),
		"-j", "ACCEPT",
	).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return err
	}

	out, err = exec.Command(
		"/usr/sbin/iptables",
		"-w", // Wait for xlock
		"-"+f, "FORWARD",
		"-s", sip,
		"-d", cip,
		"-p", "tcp",
		"--sport", strconv.FormatUint(uint64(p), 10),
		"-j", "ACCEPT",
	).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return err
	}

	return nil
}

func resolveDockerIp(n string) (string, error) {
	if Docker == nil {
		var err error
		Docker, err = docker.NewClient(*DockerUrl)

		if err != nil {
			return "", err
		}

		if err := Docker.Ping(); err != nil {
			return "", err
		}
	}

	var (
		c   *docker.Container
		err error
	)

	for c == nil || c.NetworkSettings.IPAddress == "" {
		c, err = Docker.InspectContainer(n)
		if err != nil && !strings.Contains(err.Error(), "No such container") {
			return "", err
		}

		fmt.Printf("IP For %s is empty, waiting...\n", n)
		time.Sleep(500 * time.Millisecond)
	}

	return c.NetworkSettings.IPAddress, nil
}
