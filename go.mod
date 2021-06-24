module github.com/je4/sftp/v2

go 1.16

replace github.com/je4/sftp/v2 => ./

require (
	github.com/blend/go-sdk v1.20210616.2
	github.com/goph/emperror v0.17.2
	github.com/gosuri/uilive v0.0.4 // indirect
	github.com/gosuri/uiprogress v0.0.1
	github.com/machinebox/progress v0.2.0
	github.com/matryer/is v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/sftp v1.13.1
	github.com/tidwall/transform v0.0.0-20201103190739-32f242e2dbde
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
)
