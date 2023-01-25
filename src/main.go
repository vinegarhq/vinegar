// User should not have to interact with this program at any time.

package main

import (
//	"os"
	"fmt"
//	"log"
	"github.com/adrg/xdg"
)

//func build_prefix() {
//}

func main() {
	fmt.Println(xdg.DataHome)
}
