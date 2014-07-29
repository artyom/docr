**docr** renders markdown documentation found in given git repository directly from git blob store (works even on bare repositories without working copy).

	Usage of docr:
	  -bind="127.0.0.1:8080": address to listen
	  -ref="HEAD": reference (HEAD, refs/heads/develop, etc.)
	  -repo="./.git": path to repository root (.git directory)

To install, run:

	go install -u -v github.com/artyom/docr

then start by running command:

	docr -repo $GOPATH/src/github.com/artyom/docr/.git

and open this link: <http://127.0.0.1:8080/README.md> in your browser, you'll
see this file.

Files with `.md` or `.markdown` suffixes are rendered to html, other files served unmodified.
