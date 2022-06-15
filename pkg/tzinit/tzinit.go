package tzinit

// Do not import "time" here!
import (
	"os"
)

// For this work, this package must be imported before "time".
func init() {
	os.Setenv("TZ", "UTC")
}
