package libvirt

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
)

func newCopier(virConn *libvirt.Libvirt, volume *libvirt.StorageVol, size uint64) func(src io.Reader) error {
	copier := func(src io.Reader) error {
		var bytesCopied int64

		// FIXME - validate behaviour
		// https://github.com/digitalocean/go-libvirt/pull/63/files#

		r, w := io.Pipe()
		if err := virConn.StorageVolUpload(*volume, r, 0, size, 0); err != nil {
			return fmt.Errorf("Error while uploading volume %s", err)
		}

		bytesCopied, err := io.Copy(w, src)
		// if we get unexpected EOF this mean that connection was closed suddently from server side
		// the problem is not on the plugin but on server hosting currupted images
		if err == io.ErrUnexpectedEOF {
			return w.CloseWithError(fmt.Errorf("Error: transfer was unexpectedly closed from the server while downloading. Please try again later or check the server hosting sources"))
		}
		if err != nil {
			return w.CloseWithError(fmt.Errorf("Error while copying source to volume %s", err))
		}

		log.Printf("%d bytes uploaded\n", bytesCopied)
		if uint64(bytesCopied) != size {
			return w.CloseWithError(fmt.Errorf("Error during volume Upload. BytesCopied: %d != %d volume.size", bytesCopied, size))
		}

		return nil
	}
	return copier
}

func timeFromEpoch(str string) time.Time {
	var s, ns int

	ts := strings.Split(str, ".")
	if len(ts) == 2 {
		ns, _ = strconv.Atoi(ts[1])
	}
	s, _ = strconv.Atoi(ts[0])

	return time.Unix(int64(s), int64(ns))
}
