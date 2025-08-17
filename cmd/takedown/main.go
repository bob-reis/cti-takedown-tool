package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	action := flag.String("action", "", "action to perform")
	ioc := flag.String("ioc", "", "indicator of compromise")
	tags := flag.String("tags", "", "comma separated list of tags")
	caseID := flag.String("case", "", "takedown case identifier")
	daemon := flag.Bool("daemon", false, "run in daemon mode")
	flag.Parse()

	_, _ = fmt.Fprintf(os.Stdout, "takedown CLI placeholder\n")
	_, _ = fmt.Fprintf(os.Stdout, "action=%s ioc=%s tags=%s case=%s daemon=%v\n", *action, *ioc, *tags, *caseID, *daemon)
}
