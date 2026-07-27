package check

import (
	"context"
	"fmt"
	"net"
	"time"
)

var ERROR_FAILED_DIAL = fmt.Errorf("Failed to dial:")

func CheckTCP(target string) (string, error) {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		return "Failed to dial: ", err
	}
	defer conn.Close()
	return "Success", nil

}
