module github.com/blues-alex/lms6000

go 1.24.1

replace uart => ../uart

require uart v0.0.0-00010101000000-000000000000

require (
	github.com/creack/goselect v0.1.2 // indirect
	go.bug.st/serial v1.6.2 // indirect
	golang.org/x/sys v0.6.0 // indirect
)
