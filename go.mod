module github.com/je4/sftp/v2

go 1.16

replace github.com/je4/sftp/v2 => ./

require (
	github.com/goph/emperror v0.17.2
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/sftp v1.13.1
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
)
