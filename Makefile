subObjects = **/*.go
objects = *.go

all: install

install: ; go install -ldflags="-w" *.go

go: ; go install -ldflags="-w" *.go
