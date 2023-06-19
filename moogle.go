package main

import (
	"fmt"

	"github.com/anicolaspp/moogle/server"
)

func main() {
	fmt.Println("Hello Moogle!")

	moogle := server.Moogle{}
	if err := moogle.Run(); err != nil {
		fmt.Printf("Error: %v running the Moogle Server...\n", err)
	}

	fmt.Println("Bye Bye Moogle!")
}
