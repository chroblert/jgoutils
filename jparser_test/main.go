package main

import (
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jparser/pe"
)

func main() {
	pefile, err := pe.Parse("f:\\artifact.exe")
	if err != nil {
		jlog.Error(err)
		return
	}
	jlog.Debug(pefile.DosHeader)
	jlog.Debug(pefile.FileHeader)
	jlog.Debug(pefile.OptionalHeader)
	jlog.Debugf("%08x,%08x", pefile.OptionalHeader.(*pe.IMAGE_OPTIONAL_HEADER64).DataDirectory[4].VirtualAddress, pefile.OptionalHeader.(*pe.IMAGE_OPTIONAL_HEADER64).DataDirectory[4].Size)

}
