package godless

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

func imin(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func writeBytes(bs []byte, w io.Writer) error {
	written, err := w.Write(bs)

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("write failed after %v bytes", written))
	}

	return nil
}
