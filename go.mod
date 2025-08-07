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
	github.com/godbus/dbus/v5 v5.1.0
	github.com/jwijenbergh/puregotk v0.0.0-20250407124134-bc1a52f44fd4
	github.com/lmittmann/tint v1.1.2
	github.com/samber/slog-multi v1.4.1
	github.com/sewnie/rbxbin v0.0.0-20250806104726-6336cb47e0e4
	github.com/sewnie/rbxweb v0.0.0-20250807144008-1cc287f8788a
	github.com/sewnie/wine v0.0.0-20250802220453-3332fe968016
	golang.org/x/sys v0.35.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jwijenbergh/purego v0.0.0-20241210143217-aeaa0bfe09e0 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	github.com/samber/lo v1.51.0 // indirect
	github.com/samber/slog-common v0.19.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)

retract (
	[v1.0.0, v1.1.3]
	v0.0.1
)

replace github.com/jwijenbergh/puregotk => github.com/sewnie/puregotk v0.0.0-20250803195448-7ce64774bae4
