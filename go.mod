module github.com/vinegarhq/vinegar

go 1.22.0

require (
	github.com/BurntSushi/toml v1.4.0
	github.com/adrg/xdg v0.5.3
	github.com/otiai10/copy v1.14.1-0.20240925044834-49b0b590f1e1
	golang.org/x/sync v0.9.0
)

require (
	gioui.org v0.7.1
	github.com/altfoxie/drpc v0.0.0-20240929140334-e714e6291275
	github.com/apprehensions/rbxbin v0.0.0-20241108182759-6d92e1ecbfab
	github.com/apprehensions/rbxweb v0.0.0-20240329184049-0bdedc184942
	github.com/apprehensions/wine v0.0.0-20241109121733-f99088878030
	github.com/folbricht/pefile v0.1.0
	github.com/fsnotify/fsnotify v1.8.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/lmittmann/tint v1.0.5
	github.com/nxadm/tail v1.4.11
	github.com/samber/slog-multi v1.2.4
	golang.org/x/sys v0.27.0
	golang.org/x/term v0.26.0
)

require (
	gioui.org/cpu v0.0.0-20220412190645-f1e9e8c3b1f7 // indirect
	gioui.org/shader v1.0.8 // indirect
	github.com/go-text/typesetting v0.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	github.com/robloxapi/rbxdhist v0.6.0 // indirect
	github.com/robloxapi/rbxver v0.3.0 // indirect
	github.com/samber/lo v1.47.0 // indirect
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0 // indirect
	golang.org/x/exp/shiny v0.0.0-20240909161429-701f63a606c0 // indirect
	golang.org/x/image v0.20.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)

retract (
	[v1.0.0, v1.1.3]
	v0.0.1
)
