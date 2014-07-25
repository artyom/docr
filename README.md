**docr** renders markdown documentation found in given git repository.

	Usage of ./docr:
	  -bind="127.0.0.1:8080": address to listen
	  -ref="HEAD": reference
	  -repo=".": path to repository root

To install, run:

	go install -u -v github.com/artyom/docr

then start by running command:

	docr -repo $GOPATH/src/github.com/artyom/docr/.git

and open this link: <http://127.0.0.1:8080/README.md> in your browser, you'll
see this file.
