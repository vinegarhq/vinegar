module github.com/vinegarhq/vinegar

go 1.23.0

toolchain go1.24.5

require (
	github.com/BurntSushi/toml v1.5.0
	github.com/adrg/xdg v0.5.3
	github.com/otiai10/copy v1.14.1
	golang.org/x/sync v0.16.0
)

require (
	github.com/altfoxie/drpc v0.0.0-20240929140334-e714e6291275
	github.com/jwijenbergh/puregotk v0.0.0-20250812133623-7203178b5172
	github.com/lmittmann/tint v1.1.2
	github.com/sewnie/rbxbin v0.0.0-20250816115250-5cf6d6f110f5
	github.com/sewnie/rbxweb v0.0.0-20250923154144-a174c75bba5d
	github.com/sewnie/wine v0.0.0-20250923132641-827c1e560ae0
	golang.org/x/sys v0.35.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jwijenbergh/purego v0.0.0-20250812133547-b5852df1402b // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)

retract (
	[v1.0.0, v1.1.3]
	v0.0.1
)

replace github.com/jwijenbergh/puregotk v0.0.0-20250812133623-7203178b5172 => github.com/sewnie/puregotk v0.0.0-20251005215301-c0269d233573
