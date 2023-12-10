package main

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"syscall/js"
)

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("archive", js.FuncOf(archive))
	<-c
}

func archive(_ js.Value, args []js.Value) any {
	if len(args) <= 0 {
		return false
	}
	files := args[0]
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	w.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})

	for i := 0; i < files.Length(); i++ {
		file := files.Index(i)
		f, err := w.Create(fmt.Sprintf("zip_result/%s", file.Get("name").String()))

		if err != nil {
			return false
		}

		goBytes := make([]byte, file.Get("value").Get("length").Int())
		js.CopyBytesToGo(goBytes, file.Get("value"))
		_, err = f.Write(goBytes)
		if err != nil {
			return false
		}
	}
	if err := w.Close(); err != nil {
		return false
	}

	return download(buf.Bytes(), "text/plain", "archive.zip")
}

func download(resultBytes []byte, mimeType string, fileName string) bool {
	u8Array := js.Global().Get("Uint8Array").New(len(resultBytes))
	js.CopyBytesToJS(u8Array, resultBytes)
	jsArray := js.Global().Get("Array").New(u8Array)

	blob := js.Global().Get("Blob").New(jsArray, js.ValueOf(map[string]interface{}{"type": mimeType}))

	aElem := js.Global().Get("document").Call("createElement", "a")
	aElem.Set("href", js.Global().Get("window").Get("URL").Call("createObjectURL", blob))
	aElem.Set("download", js.ValueOf(fileName))
	aElem.Call("click")
	outputTimestampLog()
	return true
}

func outputTimestampLog() {
	js.Global().Get("console").Call("log", js.Global().Get("Date").Call("now"))
}
