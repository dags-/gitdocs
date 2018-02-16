package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gohugoio/hugo/hugolib"
	"gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
	"gopkg.in/src-d/go-git.v4"
)

var (
	pulls    = make(chan bool)
	webPort  = flag.Int("web", 0, "Web Port")
	hookPort = flag.Int("hook", 0, "Webhook Port")
	secret   = flag.String("secret", "", "Webhook secret")
	url      = flag.String("url", "", "Github repo url")
)

func main() {
	flag.Parse()
	fmt.Println(hugolib.CommitHash)

	go handleWebhook(*secret, *hookPort)
	go handlePulls(*url)

	pulls <- true

	cmd := exec.Command("hugo", "server", "-s", "site", "-p", strconv.Itoa(*webPort))
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	fmt.Println(err)
}

func handlePulls(url string) {
	r, err := git.PlainClone("./site", false, &git.CloneOptions{URL: url, Progress: os.Stdout})
	if err == git.ErrRepositoryAlreadyExists {
		r, err = git.PlainOpen("./site")
		if err != nil {
			panic(err)
		}
	}

	for {
		<-pulls

		w, err := r.Worktree()
		if err != nil {
			panic(err)
		}

		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil {
			fmt.Println("Error pulling remote:", err)
		}

		time.Sleep(5 * time.Second)
	}
}

func handleWebhook(secret string, port int) {
	hook := github.New(&github.Config{Secret: secret})
	hook.RegisterEvents(handleCommit, github.PushEvent)
	err := webhooks.Run(hook, ":"+strconv.Itoa(port), "/webhooks")
	if err != nil {
		panic(err)
	}
}

func handleCommit(payload interface{}, header webhooks.Header) {
	go func() {
		fmt.Println("payload received:", payload)
		pulls <- true
	}()
}
