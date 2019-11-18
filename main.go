package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// It'd be cool to pipe the mermaid input file directly into the cli from memory:
// https://terinstock.com/post/2018/10/memfd_create-Temporary-in-memory-files-with-Go-and-Linux/

func execute(inputFileAbsolutePath string) string {
	// try-mermaid ❯ ./node_modules/.bin/mmdc -i flowchart.mmd -o flowchart.png
	outputFileAbsolutePath := trimExtension(inputFileAbsolutePath) + ".png"

	out, err := exec.Command("./node_modules/.bin/mmdc", "-i", inputFileAbsolutePath, "-o", outputFileAbsolutePath).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Println("Command Successfully Executed")
	fmt.Printf("Generated: %s", outputFileAbsolutePath)
	output := string(out[:])
	fmt.Println(output)
	return outputFileAbsolutePath
}

func trimExtension(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}

func getVersion() string {
	//try-mermaid ❯ ./node_modules/.bin/mmdc --version
	out, err := exec.Command("./node_modules/.bin/mmdc", "--version").Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Println("Command Successfully Executed")
	return strings.TrimSpace(string(out[:]))
}
func versionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Mermaid version: %s", getVersion())
}

func textHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		rBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		tmpFile, err := ioutil.TempFile(os.TempDir(), "mermaid-*.mmd")
		if err != nil {
			log.Fatal("Cannot create temporary file", err)
		}
		defer os.Remove(tmpFile.Name())

		fmt.Println("Created File: " + tmpFile.Name())

		if _, err = tmpFile.Write(rBody); err != nil {
			log.Fatal("Failed to write to temporary file", err)
		}

		resultAbsolutePath := execute(tmpFile.Name())
		defer os.Remove(resultAbsolutePath)

		imageBytes, err := ioutil.ReadFile(resultAbsolutePath)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(imageBytes)))
		if _, err := w.Write(imageBytes); err != nil {
			log.Println("unable to write image.")
		}

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

func main() {
	http.HandleFunc("/version", versionHandler)
	http.HandleFunc("/diag", textHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
