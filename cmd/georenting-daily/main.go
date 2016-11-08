package main

import "github.com/alternaDev/georenting-server/jobs"

func main() {
	jobs.QueueDeployGCAidRequest()
}
