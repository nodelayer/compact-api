package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/nodelayer/compact-api/internal/node"
)

type Server struct {
	RouteVersions *regexp.Regexp
	RouteLayerGen *regexp.Regexp
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"%s [?%s] [%s | %+v]\n",
		r.URL.Path,
		r.URL.RawQuery,
		r.RemoteAddr,
		strings.Join(r.Header.Values("X-Forwarded-For"), ", "),
	)

	hdr := w.Header()

	hdr.Set("Access-Control-Allow-Origin", "*")
	hdr.Set("Access-Control-Allow-Headers", "*")
	hdr.Set("Access-Control-Allow-Methods", "HEAD, GET")

	if r.Method == http.MethodGet {
		if s.RouteVersions.MatchString(r.URL.Path) {
			hdr.Set("Content-Type", "text/plain")
			fmt.Fprint(w, strings.Join(node.Versions(), "\n"))
			return
		}

		if s.RouteLayerGen.MatchString(r.URL.Path) {
			q := r.URL.Query()

			vers := q.Get("version")
			pkgs := q.Get("packages")

			if node.VersionRegexp.MatchString(vers) == false {
				hdr.Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusUnprocessableEntity)

				fmt.Fprint(w, "Unsupported version provided, see supported versions at 'GET /versions'")
				return
			}

			this := os.Args[0]
			if path.IsAbs(this) == false {
				if pwd, e := os.Getwd(); e == nil {
					this = path.Join(pwd, this)
				}
			}

			out, e := exec.Command(this, "--version", vers, "--packages", pkgs).CombinedOutput()
			if e != nil {
				msg := e.Error()
				if len(out) > 0 {
					msg = string(out)
				}

				hdr.Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusInternalServerError)

				fmt.Fprint(w, msg)
				return
			}

			lPath := strings.TrimSpace(string(out))

			f, e := os.Open(lPath)

			if e != nil {
				hdr.Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusInternalServerError)

				fmt.Fprint(w, "An unexpected error occurred, please try again later")
				return
			}

			defer f.Close()

			hdr.Set("Content-Type", "application/octet-stream")
			hdr.Set("Content-Disposition", `attachment; filename="layer.zip"`)
			w.WriteHeader(http.StatusCreated)

			io.Copy(w, f)
			return
		}
	}

	hdr.Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)

	fmt.Fprint(w, "Not Found")
}

func main() {
	if len(os.Args) <= 1 {
		srv := &http.Server{
			Addr: ":1923",
			Handler: &Server{
				RouteVersions: regexp.MustCompile(`(?i)^(\/[a-z0-9]+)*\/versions\/?$`),
				RouteLayerGen: regexp.MustCompile(`(?i)^(\/[a-z0-9]+)*\/layers\/generate\/?$`),
			},
		}

		log.Fatalln(srv.ListenAndServe())

		return
	}

	vers := flag.String("version", node.DefaultVersion, "Nodejs version to use")
	pkgs := flag.String("packages", "", "Comma-separated npm packages to install")

	flag.Parse()

	c, e := node.New(*vers)

	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
		return
	}

	npmPkgs := []string{}
	for _, pkg := range strings.Split(*pkgs, ",") {
		npmPkgs = append(npmPkgs, strings.TrimSpace(pkg))
	}

	e = c.InstallPackages(npmPkgs)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
		return
	}

	layerPath, e := c.CreateArchive()
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
		return
	}

	fmt.Fprintln(os.Stdout, layerPath)
}
