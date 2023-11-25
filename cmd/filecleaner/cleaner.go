package main

import "recorder/internal/cleaning"

func main() {
	cleaning.Clean_service(60, 24)
}
