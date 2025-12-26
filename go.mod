module github.com/vinegarhq/vinegar

go 1.24.0

toolchain go1.24.5

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/adrg/xdg v0.5.3
	github.com/otiai10/copy v1.14.1
	golang.org/x/sync v0.19.0
)

require (
	github.com/altfoxie/drpc v0.0.0-20251029175103-30d5f68745fb
	github.com/google/go-github/v80 v80.0.0
	github.com/jwijenbergh/puregotk v0.0.0-20251201161753-28ec1479c381
	github.com/lmittmann/tint v1.1.2
	github.com/sewnie/rbxbin v0.0.0-20251014195941-5c5a3767d780
	github.com/sewnie/rbxweb v0.0.0-20250923154144-a174c75bba5d
	github.com/sewnie/wine v0.0.0-20251226171014-ed3310451469
	github.com/ulikunitz/xz v0.5.15
)

require (
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jwijenbergh/purego v0.0.0-20251017112123-b71757b9ba42 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	golang.org/x/sys v0.39.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)

retract (
	[v1.0.0, v1.1.3]
	v0.0.1
)

replace github.com/altfoxie/drpc => github.com/sewnie/drpc v0.0.0-20251027131846-60568f62ffb3
