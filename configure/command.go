package configure

import "fmt"

type Options struct{}

func (o Options) Run() error {
	fmt.Println("Configuring wt...")
	return nil
}
