module github.com/vinegarhq/vinegar

go 1.22.0

require (
	github.com/BurntSushi/toml v1.4.0
	github.com/adrg/xdg v0.5.3
	github.com/otiai10/copy v1.14.1
	golang.org/x/sync v0.10.0
)

require (
	github.com/altfoxie/drpc v0.0.0-20240929140334-e714e6291275
	github.com/apprehensions/rbxbin v0.0.0-20241108182759-6d92e1ecbfab
	github.com/apprehensions/rbxweb v0.0.0-20240329184049-0bdedc184942
	github.com/apprehensions/wine v0.0.0-20250108060616-9d416603ffde
	github.com/folbricht/pefile v0.1.0
	github.com/fsnotify/fsnotify v1.8.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/jwijenbergh/puregotk v0.0.0-20240827133221-51f7e663a5e9
	github.com/lmittmann/tint v1.0.6
	github.com/nxadm/tail v1.4.11
	github.com/samber/slog-multi v1.3.3
	golang.org/x/sys v0.29.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jwijenbergh/purego v0.0.0-20240527152122-d7876464a49c // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	github.com/robloxapi/rbxdhist v0.6.0 // indirect
	github.com/robloxapi/rbxver v0.3.0 // indirect
	github.com/samber/lo v1.47.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)

retract (
	[v1.0.0, v1.1.3]
	v0.0.1
)
