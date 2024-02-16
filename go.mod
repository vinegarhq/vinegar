module github.com/vinegarhq/vinegar

go 1.21

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/adrg/xdg v0.4.0
	github.com/otiai10/copy v1.14.0
	golang.org/x/sync v0.5.0
)

require (
	gioui.org v0.4.1
	github.com/altfoxie/drpc v0.0.0-20231214171500-0a4e3a3b1c53
	github.com/folbricht/pefile v0.1.0
	github.com/fsnotify/fsnotify v1.7.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/nxadm/tail v1.4.11
	golang.org/x/sys v0.16.0
	golang.org/x/term v0.16.0
)

require (
	gioui.org/cpu v0.0.0-20220412190645-f1e9e8c3b1f7 // indirect
	gioui.org/shader v1.0.8 // indirect
	github.com/go-text/typesetting v0.0.0-20231206174126-ce41cc83e028 // indirect
	github.com/google/uuid v1.5.0 // indirect
	golang.org/x/exp v0.0.0-20231206192017-f3f8817b8deb // indirect
	golang.org/x/exp/shiny v0.0.0-20231206192017-f3f8817b8deb // indirect
	golang.org/x/image v0.14.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract (
	[v1.0.0, v1.1.3]
	v0.0.1
)
