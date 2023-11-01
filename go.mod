module github.com/vinegarhq/vinegar

go 1.21

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/adrg/xdg v0.4.0
	github.com/otiai10/copy v1.14.0
	golang.org/x/sync v0.4.0
)

require (
	dario.cat/mergo v1.0.0
	gioui.org v0.3.1
	github.com/hugolgst/rich-go v0.0.0-20230917173849-4a4fb1d3c362
	golang.org/x/sys v0.13.0
)

require (
	gioui.org/cpu v0.0.0-20220412190645-f1e9e8c3b1f7 // indirect
	gioui.org/shader v1.0.8 // indirect
	github.com/go-text/typesetting v0.0.0-20231101082850-a36c1d9288f6 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/exp/shiny v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/image v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)

retract (
	[v1.0.0, v1.1.3]
	v0.0.1
)
