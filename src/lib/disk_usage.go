package lib

import (
	"fmt"

	"github.com/ricochet2200/go-disk-usage/du"
)

func GetDiskUsage(dir string) error {
	usage := du.NewDiskUsage(dir)

	if usage == nil {
		return fmt.Errorf("error getting disk usage, usage is nil")
	}

	fmt.Printf("Disk usage for %s:\n", dir)
	fmt.Printf("Size: %d bytes\n", usage.Size())
	fmt.Printf("Free: %d bytes\n", usage.Free())
	fmt.Printf("Available: %d bytes\n", usage.Available())
	fmt.Printf("Used: %d bytes\n", usage.Used())

	return nil
}
