build:
	go mod download
	go build -o recmd-dmn

default: build

test:
	go test -v
	(cd dmn; go test -v)

clean:
	rm -f recmd-dmn

install:
	go install