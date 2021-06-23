package main

import (
	"flag"
	"fmt"
	xsftp "github.com/je4/sftp/v2/pkg/sftp"
	"net/url"
	"os"
	"regexp"
)

const (
	loglevel  = "DEBUG"
	logFormat = "%{time:2006-01-02T15:04:05.000} %{module}::%{shortfunc} [%{shortfile}] > %{level:.5s} - %{message}"
)

var targetRegex = regexp.MustCompile(`^([^@]+)@([^/:]+):([0-9]+)/(.+)$`)

func main() {
	var privateKey string
	flag.StringVar(&privateKey, "identity", "", "private key file")
	flag.Parse()
	tail := flag.Args()
	if len(tail) < 2 {
		fmt.Println("invalid parameters")
		fmt.Printf("%s -identity [private key file] [sourcepath] [sftp target]\n", os.Args[0])
		os.Exit(1)
	}
	fi, err := os.Stat(privateKey)
	if err != nil {
		fmt.Printf("cannot stat %s: %v\n", privateKey, err)
		os.Exit(1)
	}
	if fi.IsDir() {
		fmt.Printf("%s is a directory\n", privateKey)
		os.Exit(1)
	}

	src := tail[0]
	fi, err = os.Stat(src)
	if err != nil {
		fmt.Printf("cannot stat %s: %v\n", src, err)
		os.Exit(1)
	}
	if fi.IsDir() {
		fmt.Printf("%s is a directory\n", src)
		os.Exit(1)
	}

	target := tail[1]
	matches := targetRegex.FindStringSubmatch(target)
	if matches == nil {
		fmt.Printf("invalid format for target %s\n", target)
		os.Exit(1)
	}
	targetUser := matches[1]
	targetHost := matches[2]
	targetPort := matches[3]
	targetPath := matches[4]

	logger, lf := CreateLogger("sftp", "", nil, loglevel, logFormat)
	defer lf.Close()

	sftp, err := xsftp.NewSFTP([]string{privateKey}, "", "", logger)
	if err != nil {
		fmt.Printf("cannot initialize sftp: %v\n", err)
		os.Exit(1)
	}

	rawurl := fmt.Sprintf("sftp://%s@%s:%s/%s", targetUser, targetHost, targetPort, targetPath)
	targetUrl, err := url.Parse(rawurl)
	if err != nil {
		fmt.Printf("cannot parse url %s: %v\n", rawurl, err)
		os.Exit(1)
	}
	//	shaSink := sha512.New()
	//	dest := io.MultiWriter(w, shaSink)

	len, err := sftp.PutFile(targetUrl, src)
	if err != nil {
		fmt.Printf("cannot upload %s -> %s: %v\n", src, targetUrl.String(), err)
		os.Exit(1)
	}
	fmt.Printf("len: %v\n", len)
}
