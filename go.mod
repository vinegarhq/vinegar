module github.com/vinegarhq/vinegar

go 1.20

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/adrg/xdg v0.4.0
	github.com/otiai10/copy v1.12.0
	golang.org/x/sync v0.3.0
)

require golang.org/x/sys v0.11.0 // indirect

retract (
	[v1.0.0, v1.1.3]
	v0.0.1
)
