package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"text/template"
	"time"
)

type scriptWriter struct {
	outFileName     string
	outputFile      *os.File
	timestampBefore time.Time
}

func (w *scriptWriter) WriteData(first bool, data []byte) {
	ts := time.Since(w.timestampBefore).Seconds()
	w.timestampBefore = time.Now()

	cm := ","
	if first {
		cm = ""
	}
	fmt.Fprintf(w.outputFile, `%s{
			tp : 1,
			ts:%f,
			dt:"%s"
		}`,
		cm, ts, base64.StdEncoding.EncodeToString(data))
}

func (w *scriptWriter) WriteSize(first bool, size winSize) {
	ts := time.Since(w.timestampBefore).Seconds()
	w.timestampBefore = time.Now()

	cm := ","
	if first {
		cm = ""
	}

	fmt.Fprintf(w.outputFile, `%s{
			tp : 2,
			ts:%f,
			cols:%d,
		    rows:%d
		}`, cm, ts, size.cols, size.rows)

}

func (w *scriptWriter) Begin(size winSize) error {

	tempOut := template.New("output")

	// Read the main html template
	binFile, err := Asset("bindata/output_header.html.in")
	if err != nil {
		return err
	}
	tempOut.Parse(string(binFile))

	// Read the rest of the template files
	for _, fileName := range []string{"xterm.js", "xterm.css"} {
		binFile, err = Asset("bindata/" + fileName)
		if err != nil {
			return err
		}
		tempOut.New(fileName).Parse(string(binFile))
	}

	w.outputFile, err = os.Create(w.outFileName)
	if err != nil {
		return err
	}

	err = tempOut.ExecuteTemplate(w.outputFile, "output", nil)
	if err != nil {
		return err
	}

	w.timestampBefore = time.Now()
	w.WriteSize(true, size)
	return nil
}

func (w *scriptWriter) End() error {
	footer, err := Asset("bindata/output_footer.html.in")

	if err != nil {
		return err
	}
	_, err = w.outputFile.Write(footer)

	if err != nil {
		// TODO: revise all the error returns
		return err
	}
	return w.outputFile.Close()
}

func (w *scriptWriter) Write(data []byte) (n int, err error) {

	w.WriteData(false, data)
	return len(data), err
}
