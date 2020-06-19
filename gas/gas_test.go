package gas

import (
	"fmt"
	"testing"
)

func TestGetPrice(t *testing.T) {
	fmt.Println(GetPrice(Fastest))
}

func TestToWei(t *testing.T) {
	fmt.Println(ToWei(410, 8))
}
