module github.com/vinegarhq/vinegar

go 1.25.0

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/adrg/xdg v0.5.3
	github.com/otiai10/copy v1.14.1
	golang.org/x/sync v0.19.0
)

require (
	codeberg.org/puregotk/puregotk v0.0.0-20260226083027-5647eb6a2bba
	github.com/altfoxie/drpc v0.0.0-20251029175103-30d5f68745fb
	github.com/google/go-github/v80 v80.0.0
	github.com/jaypipes/pcidb v1.1.1
	github.com/lmittmann/tint v1.1.2
	github.com/pojntfx/go-gettext v0.4.0
	github.com/sewnie/rbxbin v0.0.0-20251228183315-8c321727936e
	github.com/sewnie/rbxweb v0.0.0-20250923154144-a174c75bba5d
	github.com/sewnie/wine v0.0.0-20260303224426-3dec0e3ed661
	github.com/ulikunitz/xz v0.5.15
)

require (
	codeberg.org/puregotk/purego v0.0.0-20260224095105-2513c838cb80 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	golang.org/x/sys v0.39.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)

retract (
	[v1.0.0, v1.1.3]
	v0.0.1
)

replace github.com/altfoxie/drpc => github.com/sewnie/drpc v0.0.0-20251027131846-60568f62ffb3
